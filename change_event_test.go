package agollo

import (
	"testing"
	"time"
	"github.com/zouyx/agollo/test"
	"fmt"
	"encoding/json"
)

func TestListenChangeEvent(t *testing.T) {
	go buildNotifyResult(t)

	event := ListenChangeEvent()
	defer clearChannel()
	changeEvent := <-event
	bytes, _ := json.Marshal(changeEvent)
	fmt.Println("event:", string(bytes))

	test.Equal(t,"application",changeEvent.Namespace)

	test.Equal(t,"string",changeEvent.Changes["string"].NewValue)
	test.Equal(t,"",changeEvent.Changes["string"].OldValue)
	test.Equal(t,ADDED,changeEvent.Changes["string"].ChangeType)

	test.Equal(t,"value1",changeEvent.Changes["key1"].NewValue)
	test.Equal(t,"",changeEvent.Changes["key2"].OldValue)
	test.Equal(t,ADDED,changeEvent.Changes["key1"].ChangeType)

	test.Equal(t,"value2",changeEvent.Changes["key2"].NewValue)
	test.Equal(t,"",changeEvent.Changes["key2"].OldValue)
	test.Equal(t,ADDED,changeEvent.Changes["key2"].ChangeType)

}

func clearChannel()  {
	notifyChan=nil
}

func buildNotifyResult(t *testing.T) {
	server := runChangeConfigResponse()
	defer server.Close()

	time.Sleep(1*time.Second)

	newAppConfig:=getTestAppConfig()
	newAppConfig.Ip=server.URL

	err:=autoSyncConfigServices(newAppConfig)
	err=autoSyncConfigServices(newAppConfig)

	test.Nil(t,err)

	config:=GetCurrentApolloConfig()

	test.Equal(t,"100004458",config.AppId)
	test.Equal(t,"default",config.Cluster)
	test.Equal(t,"application",config.NamespaceName)
	test.Equal(t,"20170430092936-dee2d58e74515ff3",config.ReleaseKey)
}