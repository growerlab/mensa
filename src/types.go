package src

import "github.com/growerlab/mensa/src/common"

const (
	_ = iota
	ActionTypePush
	ActionTypePull
)

var CommandActionMap = map[string]int{
	GitReceivePack:   ActionTypePush,
	GitUploadPack:    ActionTypePull,
	GitUploadArchive: ActionTypePull,
}

// 入口
// 	当用户连接到服务
type Entryer interface {
	// 进入前的预备操作
	Prep(ctx *common.Context) (err error)
	// 当进入失败时，应返回http错误码
	HttpStatus() int
	// 当进入失败时，应返回http错误的信息
	HttpStatusMessage() string
	// 错误码
	LastErr() error
}
