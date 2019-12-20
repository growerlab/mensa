package common

const (
	_               = iota
	ErrAbort        // 中间件终止
	ErrNoPermission // 无操作权限
)
