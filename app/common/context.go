package common

import (
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/growerlab/mensa/app/conf"
	"github.com/pkg/errors"
)

// 协议类型
type ProtType string

const (
	ProtTypeHTTP ProtType = "http"
	ProtTypeSSH  ProtType = "ssh"
	ProtTypeGIT  ProtType = "git"
)

const (
	ActionTypeRead  = "READ"
	ActionTypeWrite = "WRITE"
)

// 操作者
type Operator struct {
	// 当是http[s]协议时，这里可能有user、pwd(密码可能是token)
	HttpUser *url.Userinfo
	// 当是ssh协议时，可能有ssh的公钥字段
	SSHPublicKey ssh.PublicKey
}

func (o *Operator) IsHttp() bool {
	return o.HttpUser != nil
}

func (o *Operator) IsEmptyUser() bool {
	if o.IsHttp() {
		if o.HttpUser.Username() == "" {
			return true
		}
		if _, set := o.HttpUser.Password(); !set {
			return true
		}
	} else {
		if len(o.SSHPublicKey.Marshal()) == 0 {
			return true
		}
	}
	return false
}

// 相关操作的上下文
type Context struct {
	// push、pull
	ActionType string
	// 推送类型（http[s]、ssh、git）
	Type ProtType
	// ssh: 原始commands
	RawCommands []string
	// http: 原始url/commands
	RawURL string
	// http: 解析后的url
	RequestURL *url.URL
	// 仓库地址中的owner字段
	RepoOwner string
	// 仓库地址中的 仓库名
	RepoName string
	// 仓库的具体地址
	RepoDir string
	// 推送人 / 拉取人
	// 	当用户提交、拉取仓库时，应该要知道这个操作者是谁
	// 	如果仓库是公共的，那么可以忽略这个操作者字段
	// 	如果仓库是私有的，那么这个字段必须有值
	//
	Operator *Operator

	// http: 请求
	Resp http.ResponseWriter
	Req  *http.Request
}

func (c *Context) IsReadAction() bool {
	return c.ActionType == ActionTypeRead
}

func (c *Context) Desc() string {
	// who do what
	// return fmt.Sprintf(c.)
	return ""
}

func BuildContextFromHTTP(w http.ResponseWriter, r *http.Request) (*Context, error) {
	uri := r.URL
	repoOwner, repoName, repoPath, err := buildRepoInfoByPath(uri.Path)
	if err != nil {
		return nil, err
	}

	actionType := ActionTypeRead
	if uri.Query().Get("service") == "git-upload-pack" {
		actionType = ActionTypeWrite
	}

	var operator *Operator = nil
	if uri.User != nil {
		operator = &Operator{
			HttpUser: uri.User,
		}
	}

	return &Context{
		ActionType: actionType,
		Type:       ProtTypeHTTP,
		RawURL:     uri.String(),
		RequestURL: uri,
		RepoOwner:  repoOwner,
		RepoName:   repoName,
		RepoDir:    repoPath, // 仓库的具体地址
		Operator:   operator,
		Resp:       w,
		Req:        r,
	}, nil
}

func BuildContextFromSSH(session ssh.Session) (*Context, error) {
	commands := session.Command()
	if len(commands) < 2 {
		return nil, errors.Errorf("%v commands is invalid", commands)
	}

	gitPath := commands[1]
	repoOwner, repoName, repoPath, err := buildRepoInfoByPath(gitPath)
	if err != nil {
		return nil, err
	}

	actionType := ActionTypeRead
	if commands[0] == "git-upload-pack" {
		actionType = ActionTypeWrite
	}

	return &Context{
		ActionType:  actionType,
		Type:        ProtTypeSSH,
		RawCommands: commands,
		RepoOwner:   repoOwner,
		RepoName:    repoName,
		RepoDir:     repoPath, // 仓库的地址
		Operator: &Operator{
			SSHPublicKey: session.PublicKey(),
		},
	}, nil
}

func buildRepoInfoByPath(path string) (repoOwner, repoName, repoPath string, err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Println("build repo info was err: ", e)
		}
	}()

	paths := strings.FieldsFunc(path, func(r rune) bool {
		return r == rune('/') || r == rune('.')
	})
	if len(paths) < 2 {
		err = errors.Errorf("invalid repo path: %s", path)
		return
	}

	repoOwner = paths[0]
	repoName = paths[1]
	repoPath = filepath.Join(conf.GetConfig().GitRepoDir, repoOwner[:2], repoName[:2], repoOwner, repoName)
	return
}
