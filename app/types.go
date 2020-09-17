package app

import (
	"github.com/growerlab/mensa/app/common"
	"github.com/growerlab/mensa/app/middleware"
)

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
	Enter(ctx *common.Context) (result *middleware.HandleResult, err error)
}
