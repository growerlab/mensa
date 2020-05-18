package app

import "github.com/growerlab/mensa/app/common"

// var CommandActionMap = map[string]string{
// 	GitReceivePack:   common.ActionTypeWrite,
// 	GitUploadPack:    common.ActionTypeRead,
// 	GitUploadArchive: common.ActionTypeRead,
//
// 	ReceivePack: common.ActionTypeWrite,
// 	UploadPack:  common.ActionTypeRead,
// }

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
