package mensa

import "github.com/growerlab/mensa/mensa/common"

// 入口
// 	当用户连接到服务
type Entryer interface {
	// 进入前的预备操作
	Prep(ctx *common.Context) error
	// 进入失败
	Fail(reason error)
}
