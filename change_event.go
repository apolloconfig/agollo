package agollo

const (
	ADDED ConfigChangeType = iota
	MODIFIED
	DELETED
)

var (
	changeListeners []ChangeListener
)

func init() {
	changeListeners=make([]ChangeListener,0)
}

//ChangeListener 监听器
type ChangeListener interface {
	//OnChange 增加变更监控
	OnChange(event *ChangeEvent)
}


//config change type
type ConfigChangeType int

//config change event
type ChangeEvent struct {
	Namespace string
	Changes   map[string]*ConfigChange
}

type ConfigChange struct {
	OldValue   string
	NewValue   string
	ChangeType ConfigChangeType
}

//AddChangeListener 增加变更监控
func AddChangeListener(listener ChangeListener)  {
	changeListeners=append(changeListeners, listener)
}

//push config change event
func pushChangeEvent(event *ChangeEvent) {
	// if channel is null ,mean no listener,don't need to push msg
	if changeListeners == nil|| len(changeListeners)==0 {
		return
	}

	for i := range changeListeners {
		listener:= changeListeners[i]
		go listener.OnChange(event)
	}
}

//create modify config change
func createModifyConfigChange(oldValue string, newValue string) *ConfigChange {
	return &ConfigChange{
		OldValue:   oldValue,
		NewValue:   newValue,
		ChangeType: MODIFIED,
	}
}

//create add config change
func createAddConfigChange(newValue string) *ConfigChange {
	return &ConfigChange{
		NewValue:   newValue,
		ChangeType: ADDED,
	}
}

//create delete config change
func createDeletedConfigChange(oldValue string) *ConfigChange {
	return &ConfigChange{
		OldValue:   oldValue,
		ChangeType: DELETED,
	}
}
