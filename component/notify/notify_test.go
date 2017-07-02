package notify

import (
	"testing"
	"github.com/devfeel/dotweb/test"
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