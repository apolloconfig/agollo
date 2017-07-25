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
	url:=GetConfigUrl(appConfig)
	test.Equal(t,"http://localhost:8888/configs/test/dev/application?releaseKey=&ip=192.168.199.214",url)
}

func TestGetNotifyUrl(t *testing.T) {
	appConfig:=getTestAppConfig()
	url:=GetNotifyUrl("notifys",appConfig)
	test.Equal(t,"http://localhost:8888/notifications/v2?appId=test&cluster=dev&notifications=notifys",url)
}

func getTestAppConfig() *AppConfig {
	jsonStr:=`{
    "appId": "test",
    "cluster": "dev",
    "namespaceName": "application",
    "ip": "localhost:8888",
    "releaseKey": ""
	}`
	config,_:=createAppConfigWithJson(jsonStr)

	return config
}