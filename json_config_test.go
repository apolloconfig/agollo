package agollo

import (
	"testing"
	"github.com/zouyx/agollo/test"
)

func TestLoadJsonConfig(t *testing.T) {
	config,err:=LoadJsonConfig(appConfigFileName)
	t.Log(config)

	test.Nil(t,err)
	test.NotNil(t,config)
	test.Equal(t,"test",config.AppId)
	test.Equal(t,"dev",config.Cluster)
	test.Equal(t,"application",config.NamespaceName)
	test.Equal(t,"localhost:8888",config.Ip)

}

func TestLoadJsonConfigWrongFile(t *testing.T) {
	config,err:=LoadJsonConfig("")
	test.NotNil(t,err)
	test.Nil(t,config)

	test.StartWith(t,"Fail to read config file",err.Error())
}

func TestLoadJsonConfigWrongType(t *testing.T) {
	config,err:=LoadJsonConfig("app_config.go")
	test.NotNil(t,err)
	test.Nil(t,config)

	test.StartWith(t,"Load Json Config fail",err.Error())
}

func TestCreateApolloConfigWithJson(t *testing.T) {
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

func TestCreateApolloConfigWithJsonWrongEnv(t *testing.T) {
	jsonStr:=`{
    "appId": "test",
    "cluster": "joe",
    "namespaceName": "application",
    "ip": "localhost:8888",
    "releaseKey": ""
	}`
	config,err:=createAppConfigWithJson(jsonStr)
	t.Log(config)
	t.Log(err)

	test.NotNil(t,err)
	test.Nil(t,config)
	test.StartWith(t,"Env is wrong ,current env:joe",err.Error())
}

func TestCreateApolloConfigWithJsonError(t *testing.T) {
	jsonStr:=`package agollo

import (
	"os"
	"strconv"
	"github.com/cihub/seelog"
	"time"
	"fmt"
	"net/url"
)`
	config,err:=createAppConfigWithJson(jsonStr)
	t.Log(err)

	test.NotNil(t,err)
	test.Nil(t,config)
}