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

	//DELETED
	test.Equal(t,"value",changeEvent.Changes["string"].OldValue)
	test.Equal(t,"string",changeEvent.Changes["string"].NewValue)
	test.Equal(t,MODIFIED,changeEvent.Changes["string"].ChangeType)

	test.Equal(t,"true",changeEvent.Changes["bool"].OldValue)
	test.Equal(t,"",changeEvent.Changes["bool"].NewValue)
	test.Equal(t,DELETED,changeEvent.Changes["bool"].ChangeType)

	test.Equal(t,"190.3",changeEvent.Changes["float"].OldValue)
	test.Equal(t,"",changeEvent.Changes["float"].NewValue)
	test.Equal(t,DELETED,changeEvent.Changes["float"].ChangeType)

	test.Equal(t,"1",changeEvent.Changes["int"].OldValue)
	test.Equal(t,"",changeEvent.Changes["int"].NewValue)
	test.Equal(t,DELETED,changeEvent.Changes["int"].ChangeType)

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
	go runMockConfigServer(changeConfigResponse)
	defer closeMockConfigServer()

	time.Sleep(1*time.Second)

	appConfig.NextTryConnTime=0

	err:=autoSyncConfigServices()
	err=autoSyncConfigServices()

	test.Nil(t,err)

	config:=GetCurrentApolloConfig()

	test.Equal(t,"100004458",config.AppId)
	test.Equal(t,"default",config.Cluster)
	test.Equal(t,"application",config.NamespaceName)
	test.Equal(t,"20170430092936-dee2d58e74515ff3",config.ReleaseKey)
}