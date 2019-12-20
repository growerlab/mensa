package common

import (
	"net/url"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/errors"
)

// 协议类型
type ProtType string

const (
	ProtTypeHTTP ProtType = "http"
	ProtTypeSSH           = "ssh"
	ProtTypeGIT           = "git"
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
	// 原始url
	RawURL string
	// 解析后的url
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
	paths := strings.FieldsFunc(uri.Path, func(r rune) bool {
		return r == rune('/') || r == rune('.')
	})
	if len(paths) < 2 {
		return nil, errors.Errorf("invalid repo url: %s", uri.String())
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
		RepoOwner:  paths[0],
		RepoName:   paths[1],
		RepoPath:   "", // TODO 仓库的具体地址
		Operator:   operator,
	}, nil
}

func BuildContextFromSSH(sshSsession ssh.Session) *Context {
	return nil
}
