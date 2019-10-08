package agollo

import (
	"encoding/json"
	"fmt"
	. "github.com/tevid/gohamcrest"
	"testing"
	"time"
)

func TestListenChangeEvent(t *testing.T) {
	go buildNotifyResult(t)

	event := ListenChangeEvent()
	defer clearChannel()
	changeEvent := <-event
	bytes, _ := json.Marshal(changeEvent)
	fmt.Println("event:", string(bytes))

	Assert(t, "application", Equal(changeEvent.Namespace))

	Assert(t, "string", Equal(changeEvent.Changes["string"].NewValue))
	Assert(t, "", Equal(changeEvent.Changes["string"].OldValue))
	Assert(t, ADDED, Equal(changeEvent.Changes["string"].ChangeType))

	Assert(t, "value1", Equal(changeEvent.Changes["key1"].NewValue))
	Assert(t, "", Equal(changeEvent.Changes["key2"].OldValue))
	Assert(t, ADDED, Equal(changeEvent.Changes["key1"].ChangeType))

	Assert(t, "value2", Equal(changeEvent.Changes["key2"].NewValue))
	Assert(t, "", Equal(changeEvent.Changes["key2"].OldValue))
	Assert(t, ADDED, Equal(changeEvent.Changes["key2"].ChangeType))

}

func clearChannel() {
	notifyChan = nil
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

	config := GetCurrentApolloConfig()

	Assert(t, "100004458", Equal(config.AppId))
	Assert(t, "default", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "20170430092936-dee2d58e74515ff3", Equal(config.ReleaseKey))
}
