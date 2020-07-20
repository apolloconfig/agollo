package storage

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
	changeListeners = list.New()
}

//ChangeListener 监听器
type ChangeListener interface {
	//OnChange 增加变更监控
	OnChange(event *ChangeEvent)

	//OnNewestChange 监控最新变更
	OnNewestChange(configuration map[string]interface{})
}

//config change type
type ConfigChangeType int

//config change event
type ChangeEvent struct {
	Namespace string
	Changes   map[string]*ConfigChange
}

type ConfigChange struct {
	OldValue   interface{}
	NewValue   interface{}
	ChangeType ConfigChangeType
}

//AddChangeListener 增加变更监控
func AddChangeListener(listener ChangeListener) {
	if listener == nil {
		return
	}
	changeListeners.PushBack(listener)
}

//RemoveChangeListener 增加变更监控
func RemoveChangeListener(listener ChangeListener) {
	if listener == nil {
		return
	}
	for i := changeListeners.Front(); i != nil; i = i.Next() {
		apolloListener := i.Value.(ChangeListener)
		if listener == apolloListener {
			changeListeners.Remove(i)
		}
	}
}

//GetChangeListeners 获取配置修改监听器列表
func GetChangeListeners() *list.List {
	return changeListeners
}

//push config change event
func pushChangeEvent(event *ChangeEvent) {
	pushChange(func(listener ChangeListener) {
		go listener.OnChange(event)
	})
}

func pushNewestChanges(configuration map[string]interface{}) {
	pushChange(func(listener ChangeListener) {
		go listener.OnNewestChange(configuration)
	})
}

func pushChange(f func(ChangeListener)) {
	// if channel is null ,mean no listener,don't need to push msg
	if changeListeners == nil || changeListeners.Len() == 0 {
		return
	}

	for i := changeListeners.Front(); i != nil; i = i.Next() {
		listener := i.Value.(ChangeListener)
		f(listener)
	}
}

//create modify config change
func createModifyConfigChange(oldValue interface{}, newValue interface{}) *ConfigChange {
	return &ConfigChange{
		OldValue:   oldValue,
		NewValue:   newValue,
		ChangeType: MODIFIED,
	}
}

//create add config change
func createAddConfigChange(newValue interface{}) *ConfigChange {
	return &ConfigChange{
		NewValue:   newValue,
		ChangeType: ADDED,
	}
}

//create delete config change
func createDeletedConfigChange(oldValue interface{}) *ConfigChange {
	return &ConfigChange{
		OldValue:   oldValue,
		ChangeType: DELETED,
	}
}

//base on changeList create Change event
func createConfigChangeEvent(changes map[string]*ConfigChange, nameSpace string) *ChangeEvent {
	return &ChangeEvent{
		Namespace: nameSpace,
		Changes:   changes,
	}
}
