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
	OldValue string
	NewValue string
	ChangeType ConfigChangeType
}

//list config change event
func ListenChangeEvent() <-chan *ChangeEvent{
	if notifyChan==nil{
		notifyChan=make(chan *ChangeEvent,1)
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
func createModifyConfigChange(oldValue string,newValue string) *ConfigChange {
	return &ConfigChange{
		OldValue:oldValue,
		NewValue:newValue,
		ChangeType:MODIFIED,
	}
}

//create add config change
func createAddConfigChange(newValue string) *ConfigChange {
	return &ConfigChange{
		NewValue:newValue,
		ChangeType:ADDED,
	}
}

//create delete config change
func createDeletedConfigChange(oldValue string) *ConfigChange {
	return &ConfigChange{
		OldValue:oldValue,
		ChangeType:DELETED,
	}
}
