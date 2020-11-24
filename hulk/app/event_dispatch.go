package app

type EventDispatcher interface {
	// 将event推送给redis的stream
	// TODO backend响应这个事件
	Dispatch(event *PushEvent) error
}

var _ EventDispatcher = (*EventDispatch)(nil)

type EventDispatch struct {
}

func (e *EventDispatch) Dispatch(event *PushEvent) error {
	panic("implement me")
}
