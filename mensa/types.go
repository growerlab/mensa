package mensa

import "github.com/growerlab/mensa/mensa/common"

// 入口
// 	当用户连接到服务
type Entryer interface {
	// 进入前的预备操作
	Prep(ctx *common.Context) (err error)
	// 当进入失败时，应返回http错误码
	HttpStatus() int
	// 当进入失败时，应返回错误的信息
	HttpStatusMessage() string
}
