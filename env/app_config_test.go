package env

import (
	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v2"
	"github.com/zouyx/agollo/v2/component"
	"github.com/zouyx/agollo/v2/utils"

	"testing"
	"time"
)

var (
	defaultNamespace = "application"
)

func TestInit(t *testing.T) {
	config := GetAppConfig(nil)
	time.Sleep(1 * time.Second)

	Assert(t, config, NotNilVal())
	Assert(t, "test", Equal(config.AppId))
	Assert(t, "dev", Equal(config.Cluster))
	Assert(t, "application,abc1", Equal(config.NamespaceName))
	Assert(t, "localhost:8888", Equal(config.Ip))

	apolloConfig := component.GetCurrentApolloConfig()[defaultNamespace]
	Assert(t, "test", Equal(apolloConfig.AppId))
	Assert(t, "dev", Equal(apolloConfig.Cluster))
	Assert(t, "application", Equal(apolloConfig.NamespaceName))

}

func TestStructInit(t *testing.T) {

	readyConfig := &AppConfig{
		AppId:         "test1",
		Cluster:       "dev1",
		NamespaceName: "application1",
		Ip:            "localhost:8889",
	}

	agollo.InitCustomConfig(func() (*AppConfig, error) {
		return readyConfig, nil
	})

	time.Sleep(1 * time.Second)

	config := GetAppConfig(nil)
	Assert(t, config, NotNilVal())
	Assert(t, "test1", Equal(config.AppId))
	Assert(t, "dev1", Equal(config.Cluster))
	Assert(t, "application1", Equal(config.NamespaceName))
	Assert(t, "localhost:8889", Equal(config.Ip))

	apolloConfig := component.GetCurrentApolloConfig()[config.NamespaceName]
	Assert(t, "test1", Equal(apolloConfig.AppId))
	Assert(t, "dev1", Equal(apolloConfig.Cluster))
	Assert(t, "application1", Equal(apolloConfig.NamespaceName))

	//revert file config
	initFileConfig()
}

func TestGetConfigUrl(t *testing.T) {
	appConfig := getTestAppConfig()
	url := getConfigUrl(appConfig)
	Assert(t, url, StartWith("http://localhost:8888/configs/test/dev/application?releaseKey=&ip="))
}

func TestGetConfigUrlByHost(t *testing.T) {
	appConfig := getTestAppConfig()
	url := getConfigUrlByHost(appConfig, "http://baidu.com/")
	Assert(t, url, StartWith("http://baidu.com/configs/test/dev/application?releaseKey=&ip="))
}

func TestGetServicesConfigUrl(t *testing.T) {
	appConfig := getTestAppConfig()
	url := getServicesConfigUrl(appConfig)
	ip := utils.GetInternal()
	Assert(t, "http://localhost:8888/services/config?appId=test&ip="+ip, Equal(url))
}

func getTestAppConfig() *AppConfig {
	jsonStr := `{
    "appId": "test",
    "cluster": "dev",
    "namespaceName": "application",
    "ip": "localhost:8888",
    "releaseKey": "1"
	}`
	config, _ := CreateAppConfigWithJson(jsonStr)

	return config
}

func TestSyncServerIpList(t *testing.T) {
	trySyncServerIpList(t)
}

func trySyncServerIpList(t *testing.T) {
	server := runMockServicesConfigServer()
	defer server.Close()

	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL
	err := syncServerIpList(newAppConfig)

	Assert(t, err, NilVal())

	serverLen := getServersLen()

	Assert(t, 10, Equal(serverLen))

}

func TestSelectHost(t *testing.T) {
	//mock ip data
	trySyncServerIpList(t)

	t.Log("appconfig host:" + appConfig.getHost())
	t.Log("appconfig select host:" + appConfig.SelectHost())

	host := "http://localhost:8888/"
	Assert(t, host, Equal(appConfig.getHost()))
	Assert(t, host, Equal(appConfig.SelectHost()))

	//check select next time
	appConfig.setNextTryConnTime(5)
	Assert(t, host, NotEqual(appConfig.SelectHost()))
	time.Sleep(6 * time.Second)
	Assert(t, host, Equal(appConfig.SelectHost()))

	//check servers
	appConfig.setNextTryConnTime(5)
	firstHost := appConfig.SelectHost()
	Assert(t, host, NotEqual(firstHost))
	SetDownNode(firstHost)

	secondHost := appConfig.SelectHost()
	Assert(t, host, NotEqual(secondHost))
	Assert(t, firstHost, NotEqual(secondHost))
	SetDownNode(secondHost)

	thirdHost := appConfig.SelectHost()
	Assert(t, host, NotEqual(thirdHost))
	Assert(t, firstHost, NotEqual(thirdHost))
	Assert(t, secondHost, NotEqual(thirdHost))

	servers.Range(func(k, v interface{}) bool {
		SetDownNode(k.(string))
		return true
	})

	Assert(t, "", Equal(appConfig.SelectHost()))

	//no servers
	//servers = make(map[string]*serverInfo, 0)
	deleteServers()
	Assert(t, "", Equal(appConfig.SelectHost()))
}

func deleteServers() {
	servers.Range(func(k, v interface{}) bool {
		servers.Delete(k)
		return true
	})
}

func getServersLen() int {
	len := 0
	servers.Range(func(k, v interface{}) bool {
		len++
		return true
	})
	return len
}
