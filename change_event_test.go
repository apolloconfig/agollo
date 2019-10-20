package agollo

import (
	"encoding/json"
	"fmt"
	. "github.com/tevid/gohamcrest"
	"sync"
	"testing"
	"time"
)

type CustomChangeListener struct {
	t *testing.T
	group *sync.WaitGroup
}

func (c *CustomChangeListener) OnChange(changeEvent *ChangeEvent) {
	defer c.group.Done()
	bytes, _ := json.Marshal(changeEvent)
	fmt.Println("event:", string(bytes))

	Assert(c.t, "application", Equal(changeEvent.Namespace))

	Assert(c.t, "string", Equal(changeEvent.Changes["string"].NewValue))
	Assert(c.t, "", Equal(changeEvent.Changes["string"].OldValue))
	Assert(c.t, ADDED, Equal(changeEvent.Changes["string"].ChangeType))

	Assert(c.t, "value1", Equal(changeEvent.Changes["key1"].NewValue))
	Assert(c.t, "", Equal(changeEvent.Changes["key2"].OldValue))
	Assert(c.t, ADDED, Equal(changeEvent.Changes["key1"].ChangeType))

	Assert(c.t, "value2", Equal(changeEvent.Changes["key2"].NewValue))
	Assert(c.t, "", Equal(changeEvent.Changes["key2"].OldValue))
	Assert(c.t, ADDED, Equal(changeEvent.Changes["key2"].ChangeType))
}

func TestListenChangeEvent(t *testing.T) {
	go buildNotifyResult(t)
	group:= sync.WaitGroup{}
	group.Add(1)

	listener := &CustomChangeListener{
		t:t,
		group:&group,
	}
	AddChangeListener(listener)
	group.Wait()
}

func buildNotifyResult(t *testing.T) {
	server := runChangeConfigResponse()
	defer server.Close()

	time.Sleep(1 * time.Second)

	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	err := autoSyncConfigServices(newAppConfig)
	err = autoSyncConfigServices(newAppConfig)

	Assert(t, err,NilVal())

	config := GetCurrentApolloConfig()[newAppConfig.NamespaceName]

	Assert(t, "100004458", Equal(config.AppId))
	Assert(t, "default", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "20170430092936-dee2d58e74515ff3", Equal(config.ReleaseKey))
}
