package agollo

import (
	"github.com/tevid/gohamcrest"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	config := GetAppConfig(nil)

	gohamcrest.Assert(t, config,gohamcrest.NotNilVal())
	gohamcrest.Assert(t, "test", gohamcrest.Equal(config.AppId))
	gohamcrest.Assert(t, "dev", gohamcrest.Equal(config.Cluster))
	gohamcrest.Assert(t, "application", gohamcrest.Equal(config.NamespaceName))
	gohamcrest.Assert(t, "localhost:8888", gohamcrest.Equal(config.Ip))

	apolloConfig := GetCurrentApolloConfig()
	gohamcrest.Assert(t, "test", gohamcrest.Equal(apolloConfig.AppId))
	gohamcrest.Assert(t, "dev", gohamcrest.Equal(apolloConfig.Cluster))
	gohamcrest.Assert(t, "application", gohamcrest.Equal(apolloConfig.NamespaceName))

}

func TestStructInit(t *testing.T) {

	readyConfig := &AppConfig{
		AppId:         "test1",
		Cluster:       "dev1",
		NamespaceName: "application1",
		Ip:            "localhost:8889",
	}

	InitCustomConfig(func() (*AppConfig, error) {
		return readyConfig, nil
	})

	config := GetAppConfig(nil)
	gohamcrest.Assert(t, config, gohamcrest.NotNilVal())
	gohamcrest.Assert(t, "test1", gohamcrest.Equal(config.AppId))
	gohamcrest.Assert(t, "dev1", gohamcrest.Equal(config.Cluster))
	gohamcrest.Assert(t, "application1", gohamcrest.Equal(config.NamespaceName))
	gohamcrest.Assert(t, "localhost:8889", gohamcrest.Equal(config.Ip))

	apolloConfig := GetCurrentApolloConfig()
	gohamcrest.Assert(t, "test1", gohamcrest.Equal(apolloConfig.AppId))
	gohamcrest.Assert(t, "dev1", gohamcrest.Equal(apolloConfig.Cluster))
	gohamcrest.Assert(t, "application1", gohamcrest.Equal(apolloConfig.NamespaceName))

	//revert file config
	initFileConfig()
}

func TestGetConfigUrl(t *testing.T) {
	appConfig := getTestAppConfig()
	url := getConfigUrl(appConfig)
	gohamcrest.Assert(t, "http://localhost:8888/configs/test/dev/application?releaseKey=&ip=", gohamcrest.StartWith(url))
}

func TestGetConfigUrlByHost(t *testing.T) {
	appConfig := getTestAppConfig()
	url := getConfigUrlByHost(appConfig, "http://baidu.com/")
	gohamcrest.Assert(t, "http://baidu.com/configs/test/dev/application?releaseKey=&ip=", gohamcrest.StartWith(url))
}

func TestGetServicesConfigUrl(t *testing.T) {
	appConfig := getTestAppConfig()
	url := getServicesConfigUrl(appConfig)
	ip := getInternal()
	gohamcrest.Assert(t, "http://localhost:8888/services/config?appId=test&ip="+ip, gohamcrest.Equal(url))
}

func getTestAppConfig() *AppConfig {
	jsonStr := `{
    "appId": "test",
    "cluster": "dev",
    "namespaceName": "application",
    "ip": "localhost:8888",
    "releaseKey": "1"
	}`
	config, _ := createAppConfigWithJson(jsonStr)

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

	gohamcrest.Assert(t, err,gohamcrest.NotNilVal())

	gohamcrest.Assert(t, 10, gohamcrest.Equal(len(servers)))

}

func TestSelectHost(t *testing.T) {
	//mock ip data
	trySyncServerIpList(t)

	t.Log("appconfig host:" + appConfig.getHost())
	t.Log("appconfig select host:" + appConfig.selectHost())

	host := "http://localhost:8888/"
	gohamcrest.Assert(t, host, gohamcrest.Equal(appConfig.getHost()))
	gohamcrest.Assert(t, host, gohamcrest.Equal(appConfig.selectHost()))

	//check select next time
	appConfig.setNextTryConnTime(5)
	gohamcrest.Assert(t, host, gohamcrest.NotEqual(appConfig.selectHost()))
	time.Sleep(6 * time.Second)
	gohamcrest.Assert(t, host, gohamcrest.Equal(appConfig.selectHost()))

	//check servers
	appConfig.setNextTryConnTime(5)
	firstHost := appConfig.selectHost()
	gohamcrest.Assert(t, host, gohamcrest.NotEqual(firstHost))
	setDownNode(firstHost)

	secondHost := appConfig.selectHost()
	gohamcrest.Assert(t, host, gohamcrest.NotEqual(secondHost))
	gohamcrest.Assert(t, firstHost, gohamcrest.NotEqual(secondHost))
	setDownNode(secondHost)

	thirdHost := appConfig.selectHost()
	gohamcrest.Assert(t, host, gohamcrest.NotEqual(thirdHost))
	gohamcrest.Assert(t, firstHost, gohamcrest.NotEqual(thirdHost))
	gohamcrest.Assert(t, secondHost, gohamcrest.NotEqual(thirdHost))

	for host := range servers {
		setDownNode(host)
	}

	gohamcrest.Assert(t, "", gohamcrest.Equal(appConfig.selectHost()))

	//no servers
	servers = make(map[string]*serverInfo, 0)
	gohamcrest.Assert(t, "", gohamcrest.Equal(appConfig.selectHost()))
}
