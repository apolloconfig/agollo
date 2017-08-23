package agollo

import (
	"testing"
	"os"
	"strconv"
	"time"
	"github.com/zouyx/agollo/test"
)

func TestInit(t *testing.T) {
	config:=GetAppConfig()

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

func TestInitRefreshInterval_1(t *testing.T) {
	os.Setenv(refresh_interval_key,"joe")

	err:=initRefreshInterval()
	test.NotNil(t,err)

	interval:="3"
	os.Setenv(refresh_interval_key,interval)
	err=initRefreshInterval()
	test.Nil(t,err)
	i,_:=strconv.Atoi(interval)
	test.Equal(t,time.Duration(i),refresh_interval)

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

func TestGetNotifyUrl(t *testing.T) {
	appConfig:=getTestAppConfig()
	url:=getNotifyUrl("notifys",appConfig)
	test.Equal(t,"http://localhost:8888/notifications/v2?appId=test&cluster=dev&notifications=notifys",url)
}

func TestGetNotifyUrlByHost(t *testing.T) {
	appConfig:=getTestAppConfig()
	url:=getNotifyUrlByHost("notifys",appConfig,"http://baidu.com/")
	test.Equal(t,"http://baidu.com/notifications/v2?appId=test&cluster=dev&notifications=notifys",url)
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
	go runMockServicesConfigServer(normalServicesConfigResponse)
	defer closeMockServicesConfigServer()

	time.Sleep(1*time.Second)

	err:=syncServerIpList()

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