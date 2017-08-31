package agollo

import (
	"testing"
	"time"

	"github.com/zouyx/agollo/test"
)

//func TestInitRefreshInterval(t *testing.T) {
//	refresh_interval=1*time.Second
//
//	var c AbsComponent
//	c=&AutoRefreshConfigComponent{}
//	c.Start()
//}

func TestAutoSyncConfigServices(t *testing.T) {
	runMockConfigServer(normalConfigResponse)
	defer closeMockConfigServer()

	time.Sleep(1 * time.Second)

	appConfig.NextTryConnTime = 0

	err := autoSyncConfigServices()
	err = autoSyncConfigServices()

	test.Nil(t, err)

	config := GetCurrentApolloConfig()

	test.Equal(t, "100004458", config.AppId)
	test.Equal(t, "default", config.Cluster)
	test.Equal(t, "application", config.NamespaceName)
	test.Equal(t, "20170430092936-dee2d58e74515ff3", config.ReleaseKey)
	//test.Equal(t,"value1",config.Configurations["key1"])
	//test.Equal(t,"value2",config.Configurations["key2"])
}

func TestAutoSyncConfigServicesError(t *testing.T) {
	//reload app properties
	go initConfig()
	go runMockConfigServer(errorConfigResponse)
	defer closeMockConfigServer()

	time.Sleep(1 * time.Second)

	err := autoSyncConfigServices()

	test.NotNil(t, err)

	config := GetCurrentApolloConfig()

	//still properties config
	test.Equal(t, "test", config.AppId)
	test.Equal(t, "dev", config.Cluster)
	test.Equal(t, "application", config.NamespaceName)
	test.Equal(t, "", config.ReleaseKey)
}
