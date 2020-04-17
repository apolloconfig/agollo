package storage

import (
	"sync"
	"testing"

	. "github.com/tevid/gohamcrest"
)

var listener *CustomChangeListener

func init() {
	listener = &CustomChangeListener{}
}

type CustomChangeListener struct {
	w sync.WaitGroup
}

func (t *CustomChangeListener) OnChange(event *ChangeEvent) {
	t.w.Done()
}

func TestAddChangeListener(t *testing.T) {

	AddChangeListener(nil)
	Assert(t, changeListeners.Len(), Equal(0))

	AddChangeListener(listener)

	Assert(t, changeListeners.Len(), Equal(1))
}

func TestGetChangeListeners(t *testing.T) {
	Assert(t, GetChangeListeners().Len(), Equal(1))
}

func TestRemoveChangeListener(t *testing.T) {
	RemoveChangeListener(nil)
	Assert(t, changeListeners.Len(), Equal(1))
	RemoveChangeListener(listener)
	Assert(t, changeListeners.Len(), Equal(0))
}

func TestPushChangeEvent(t *testing.T) {

	addConfig := createAddConfigChange("new")
	deleteConfig := createDeletedConfigChange("old")
	modifyConfig := createModifyConfigChange("old", "new")
	changes := make(map[string]*ConfigChange)
	changes["add"] = addConfig
	event := &ChangeEvent{
		"a",
		changes,
	}
	changes["delete"] = deleteConfig
	changes["modify"] = modifyConfig
	listener = &CustomChangeListener{}
	listener.w.Add(1)

	AddChangeListener(listener)

	pushChangeEvent(event)

	listener.w.Wait()

	RemoveChangeListener(listener)
}

func TestCreateConfigChangeEvent(t *testing.T) {
	addConfig := createAddConfigChange("new")
	changes := make(map[string]*ConfigChange)
	changes["add"] = addConfig
	event := createConfigChangeEvent(changes, "ns")
	Assert(t, event, NotNilVal())
	Assert(t, len(event.Changes), Equal(1))
	Assert(t, event.Namespace, Equal("ns"))
}
