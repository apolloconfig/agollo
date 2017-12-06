package agollo

const(
	ADDED ConfigChangeType=iota
	MODIFIED
	DELETED
)

var(
	notifyChan=make(chan *ChangeEvent)
)

//状态改变类型
type ConfigChangeType int

//配置改变事件
type ChangeEvent struct {
	Namespace string
	Changes map[string]ConfigChange
}

type ConfigChange struct {
	Key string
	OldValue string
	NewValue string
	ChangeType ConfigChangeType
}

//监听配置改变事件
func ListenChangeEvent() chan *ChangeEvent{
	return notifyChan
}

//推送配置改变事件
func PushChangeEvent(event *ChangeEvent) {
	notifyChan<-event
}
