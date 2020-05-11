package common

import (
	"log"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/growerlab/mensa/src/conf"
	"github.com/pkg/errors"
)

// 协议类型
type ProtType string

const (
	ProtTypeHTTP ProtType = "http"
	ProtTypeSSH  ProtType = "ssh"
	ProtTypeGIT  ProtType = "git"
)

// 操作者
type Operator struct {
	// 当是http[s]协议时，这里可能有user、pwd(密码可能是token)
	HttpUser *url.Userinfo
	// 当是ssh协议时，可能有ssh的公钥字段
	SSHPublicKey ssh.PublicKey
}

// 相关操作的上下文
type Context struct {
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
	RepoPath string
	// 推送人 / 拉取人
	// 	当用户提交、拉取仓库时，应该要知道这个操作者是谁
	// 	如果仓库是公共的，那么可以忽略这个操作者字段
	// 	如果仓库是私有的，那么这个字段必须有值
	//
	Operator *Operator
}

func BuildContextFromHTTP(uri *url.URL) (*Context, error) {
	repoOwner, repoName, repoPath, err := buildRepoInfoByPath(uri.Path)
	if err != nil {
		return nil, err
	}

	var operator *Operator = nil
	if uri.User != nil {
		operator = &Operator{
			HttpUser: uri.User,
		}
	}

	return &Context{
		Type:       ProtTypeHTTP,
		RawURL:     uri.String(),
		RequestURL: uri,
		RepoOwner:  repoOwner,
		RepoName:   repoName,
		RepoPath:   repoPath, // 仓库的具体地址
		Operator:   operator,
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

	return &Context{
		Type:        ProtTypeSSH,
		RawCommands: commands,
		RepoOwner:   repoOwner,
		RepoName:    repoName,
		RepoPath:    repoPath, // 仓库的地址
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
