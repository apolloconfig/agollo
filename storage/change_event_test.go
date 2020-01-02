package storage

import (
	. "github.com/tevid/gohamcrest"
	"sync"
	"testing"
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

	AddChangeListener(listener)

	Assert(t, changeListeners.Len(), Equal(1))
}

func TestGetChangeListeners(t *testing.T) {
	Assert(t, GetChangeListeners().Len(), Equal(1))
}

func TestRemoveChangeListener(t *testing.T) {
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
}
