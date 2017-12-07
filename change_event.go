package agollo

const(
	ADDED ConfigChangeType=iota
	MODIFIED
	DELETED
)

var(
	notifyChan chan *ChangeEvent
)

//config change type
type ConfigChangeType int

//config change event
type ChangeEvent struct {
	Namespace string
	Changes map[string]*ConfigChange
}

type ConfigChange struct {
	Key string
	OldValue string
	NewValue string
	ChangeType ConfigChangeType
}

//list config change event
func ListenChangeEvent() chan *ChangeEvent{
	if notifyChan==nil{
		notifyChan=make(chan *ChangeEvent,10)
	}
	return notifyChan
}

//push config change event
func pushChangeEvent(event *ChangeEvent) {
	// if channel is null ,mean no listener,don't need to push msg
	if notifyChan==nil{
		return
	}

	notifyChan<-event
}

//create modify config change
func createModifyConfigChange(key string,oldValue string,newValue string) *ConfigChange {
	return &ConfigChange{
		Key:key,
		OldValue:oldValue,
		NewValue:newValue,
		ChangeType:MODIFIED,
	}
}

//create add config change
func createAddConfigChange(key string,newValue string) *ConfigChange {
	return &ConfigChange{
		Key:key,
		NewValue:newValue,
		ChangeType:ADDED,
	}
}

//create delete config change
func createDeletedConfigChange(key string,oldValue string) *ConfigChange {
	return &ConfigChange{
		Key:key,
		OldValue:oldValue,
		ChangeType:DELETED,
	}
}
