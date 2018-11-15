package agollo

import (
	"github.com/zouyx/agollo/test"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	config:=GetAppConfig(nil)

	test.NotNil(t,config)
	test.Equal(t,"test",config.AppId)
	test.Equal(t,"dev",config.Cluster)
	test.Equal(t,"application",config.NamespaceName)
	test.Equal(t,"localhost:8888",config.Ip)

	apolloConfig:=GetCurrentApolloConfig()
	test.Equal(t,"test",apolloConfig.AppId)
	test.Equal(t,"dev",apolloConfig.Cluster)
	test.Equal(t,"application",apolloConfig.NamespaceName)

}

func TestStructInit(t *testing.T) {

	readyConfig:=&AppConfig{
		AppId:"test1",
		Cluster:"dev1",
		NamespaceName:"application1",
		Ip:"localhost:8889",
	}

	InitCustomConfig(func() (*AppConfig, error) {
		return readyConfig,nil
	})

	config:=GetAppConfig(nil)
	test.NotNil(t,config)
	test.Equal(t,"test1",config.AppId)
	test.Equal(t,"dev1",config.Cluster)
	test.Equal(t,"application1",config.NamespaceName)
	test.Equal(t,"localhost:8889",config.Ip)

	apolloConfig:=GetCurrentApolloConfig()
	test.Equal(t,"test1",apolloConfig.AppId)
	test.Equal(t,"dev1",apolloConfig.Cluster)
	test.Equal(t,"application1",apolloConfig.NamespaceName)

	//revert file config
	initFileConfig()
}

func TestGetConfigUrl(t *testing.T) {
	appConfig:=getTestAppConfig()
	url:=getConfigUrl(appConfig)
	test.StartWith(t,"http://localhost:8888/configs/test/dev/application?releaseKey=&ip=",url)
}

func TestGetConfigUrlByHost(t *testing.T) {
	appConfig:=getTestAppConfig()
	url:=getConfigUrlByHost(appConfig,"http://baidu.com/")
	test.StartWith(t,"http://baidu.com/configs/test/dev/application?releaseKey=&ip=",url)
}

func TestGetServicesConfigUrl(t *testing.T) {
	appConfig:=getTestAppConfig()
	url:=getServicesConfigUrl(appConfig)
	ip:=getInternal()
	test.Equal(t,"http://localhost:8888/services/config?appId=test&ip="+ip,url)
}

func getTestAppConfig() *AppConfig {
	jsonStr:=`{
    "appId": "test",
    "cluster": "dev",
    "namespaceName": "application",
    "ip": "localhost:8888",
    "releaseKey": "1"
	}`
	config,_:=createAppConfigWithJson(jsonStr)

	return config
}

func TestSyncServerIpList(t *testing.T) {
	trySyncServerIpList(t)
}

func trySyncServerIpList(t *testing.T) {
	server := runMockServicesConfigServer()
	defer server.Close()

	newAppConfig:=getTestAppConfig()
	newAppConfig.Ip=server.URL
	err:=syncServerIpList(newAppConfig)

	test.Nil(t,err)

	test.Equal(t,10,len(servers))

}

func TestSelectHost(t *testing.T) {
	//mock ip data
	trySyncServerIpList(t)

	t.Log("appconfig host:"+appConfig.getHost())
	t.Log("appconfig select host:"+appConfig.selectHost())

	host:="http://localhost:8888/"
	test.Equal(t,host,appConfig.getHost())
	test.Equal(t,host,appConfig.selectHost())


	//check select next time
	appConfig.setNextTryConnTime(5)
	test.NotEqual(t,host,appConfig.selectHost())
	time.Sleep(6*time.Second)
	test.Equal(t,host,appConfig.selectHost())

	//check servers
	appConfig.setNextTryConnTime(5)
	firstHost:=appConfig.selectHost()
	test.NotEqual(t,host,firstHost)
	setDownNode(firstHost)

	secondHost:=appConfig.selectHost()
	test.NotEqual(t,host,secondHost)
	test.NotEqual(t,firstHost,secondHost)
	setDownNode(secondHost)

	thirdHost:=appConfig.selectHost()
	test.NotEqual(t,host,thirdHost)
	test.NotEqual(t,firstHost,thirdHost)
	test.NotEqual(t,secondHost,thirdHost)


	for host,_:=range servers{
		setDownNode(host)
	}

	test.Equal(t,"",appConfig.selectHost())

	//no servers
	servers=make(map[string]*serverInfo,0)
	test.Equal(t,"",appConfig.selectHost())
}