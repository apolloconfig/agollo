package agollo

import (
	"container/list"
	"encoding/json"
	"fmt"
	"github.com/zouyx/agollo/v2/component/notify"
	"sync"
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v2/component"
)

type CustomChangeListener struct {
	t     *testing.T
	group *sync.WaitGroup
}

func (c *CustomChangeListener) OnChange(changeEvent *ChangeEvent) {
	if c.group == nil {
		return
	}
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
	group := sync.WaitGroup{}
	group.Add(1)

	listener := &CustomChangeListener{
		t:     t,
		group: &group,
	}
	AddChangeListener(listener)
	group.Wait()
	//运行完清空变更队列
	changeListeners = list.New()
}

func buildNotifyResult(t *testing.T) {
	server := runChangeConfigResponse()
	defer server.Close()

	time.Sleep(1 * time.Second)

	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	err := notify.AutoSyncConfigServices(newAppConfig)
	err = notify.AutoSyncConfigServices(newAppConfig)

	Assert(t, err, NilVal())

	config := component.GetCurrentApolloConfig()[newAppConfig.NamespaceName]

	Assert(t, "100004458", Equal(config.AppId))
	Assert(t, "default", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "20170430092936-dee2d58e74515ff3", Equal(config.ReleaseKey))
}

func TestRemoveChangeListener(t *testing.T) {
	go buildNotifyResult(t)

	listener := &CustomChangeListener{}
	AddChangeListener(listener)
	Assert(t, 1, Equal(changeListeners.Len()))
	removeChangeListener(listener)
	Assert(t, 0, Equal(changeListeners.Len()))

	//运行完清空变更队列
	changeListeners = list.New()
}
