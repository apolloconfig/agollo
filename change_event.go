package agollo

import (
	"container/list"
)

const (
	ADDED ConfigChangeType = iota
	MODIFIED
	DELETED
)

var (
	changeListeners *list.List
)

func init() {
	changeListeners=list.New()
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
	if listener==nil{
		return
	}
	changeListeners.PushBack(listener)
}

//RemoveChangeListener 增加变更监控
func removeChangeListener(listener ChangeListener)  {
	if listener==nil{
		return
	}
	for i := changeListeners.Front(); i != nil; i = i.Next() {
		apolloListener:= i.Value.(ChangeListener)
		if listener==apolloListener{
			changeListeners.Remove(i)
		}
	}
}

//push config change event
func pushChangeEvent(event *ChangeEvent) {
	// if channel is null ,mean no listener,don't need to push msg
	if changeListeners == nil||changeListeners.Len()==0 {
		return
	}

	for i := changeListeners.Front(); i != nil; i = i.Next() {
		listener:= i.Value.(ChangeListener)
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
