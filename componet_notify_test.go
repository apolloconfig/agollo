package agollo

import (
	"testing"
	"time"
	"fmt"
	"github.com/zouyx/agollo/test"
)
func TestSyncConfigServices(t *testing.T) {
	syncConfigServices()
}


func TestGetRemoteConfig(t *testing.T) {
	remoteConfigs,err:=getRemoteConfig()

	test.Nil(t,err)
	test.NotNil(t,remoteConfigs)
	t.Log("remoteConfigs:",remoteConfigs)
	t.Log("remoteConfigs size:",len(remoteConfigs))
}

func TestNotifyConfigComponent(t *testing.T) {
	go func() {
		for{
			time.Sleep(5*time.Second)
			fmt.Println(GetCurrentApolloConfig())
		}
	}()


	c:=&NotifyConfigComponent{}
	c.Start()

}