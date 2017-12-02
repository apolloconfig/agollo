package agollo

import (
	"testing"
	"github.com/zouyx/agollo/test"
)

func TestLoadJsonConfig(t *testing.T) {
	config,err:=loadJsonConfig(appConfigFileName)
	t.Log(config)

	test.Nil(t,err)
	test.NotNil(t,config)
	test.Equal(t,"test",config.AppId)
	test.Equal(t,"dev",config.Cluster)
	test.Equal(t,"application",config.NamespaceName)
	test.Equal(t,"localhost:8888",config.Ip)

}

func TestLoadJsonConfigWrongFile(t *testing.T) {
	config,err:=loadJsonConfig("")
	test.NotNil(t,err)
	test.Nil(t,config)

	test.StartWith(t,"Fail to read config file",err.Error())
}

func TestLoadJsonConfigWrongType(t *testing.T) {
	config,err:=loadJsonConfig("app_config.go")
	test.NotNil(t,err)
	test.Nil(t,config)

	test.StartWith(t,"Load Json Config fail",err.Error())
}

func TestCreateAppConfigWithJson(t *testing.T) {
	jsonStr:=`{
    "appId": "test",
    "cluster": "dev",
    "namespaceName": "application",
    "ip": "localhost:8888",
    "releaseKey": ""
	}`
	config,err:=createAppConfigWithJson(jsonStr)
	t.Log(config)

	test.Nil(t,err)
	test.NotNil(t,config)
	test.Equal(t,"test",config.AppId)
	test.Equal(t,"dev",config.Cluster)
	test.Equal(t,"application",config.NamespaceName)
	test.Equal(t,"localhost:8888",config.Ip)
}

//func TestCreateAppConfigWithJsonWrongEnv(t *testing.T) {
//	jsonStr:=`{
//    "appId": "test",
//    "cluster": "joe",
//    "namespaceName": "application",
//    "ip": "localhost:8888",
//    "releaseKey": ""
//	}`
//	config,err:=createAppConfigWithJson(jsonStr)
//	t.Log(config)
//	t.Log(err)
//
//	test.NotNil(t,err)
//	test.Nil(t,config)
//	test.StartWith(t,"Env is wrong ,current env:joe",err.Error())
//}

func TestCreateAppConfigWithJsonError(t *testing.T) {
	jsonStr:=`package agollo

import (
	"os"
	"strconv"
	"time"
	"fmt"
	"net/url"
)`
	config,err:=createAppConfigWithJson(jsonStr)
	t.Log(err)

	test.NotNil(t,err)
	test.Nil(t,config)
}

func TestCreateAppConfigWithJsonDefault(t *testing.T) {
	jsonStr:=`{
    "appId": "testDefault",
    "ip": "localhost:9999"
	}`
	config,err:=createAppConfigWithJson(jsonStr)
	t.Log(err)

	test.Nil(t,err)
	test.NotNil(t,config)
	test.Equal(t,"testDefault",config.AppId)
	test.Equal(t,"default",config.Cluster)
	test.Equal(t,"application",config.NamespaceName)
	test.Equal(t,"localhost:9999",config.Ip)
}