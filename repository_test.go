package agollo

import (
	"testing"
	"github.com/zouyx/agollo/test"
	"time"
	"encoding/json"
)

//init param
func init()  {
	//wait 1s for another go routine update apollo config
	time.Sleep(1*time.Second)

	createMockApolloConfig(configCacheExpireTime)
}

func createMockApolloConfig(expireTime int)map[string]string{
	configs:=make(map[string]string,0)
	//string
	configs["string"]="value"
	//int
	configs["int"]="1"
	//float
	configs["float"]="190.3"
	//bool
	configs["bool"]="true"

	updateApolloConfigCache(configs,expireTime)

	return configs
}

func TestUpdateApolloConfigNull(t *testing.T) {
	time.Sleep(1*time.Second)
	var currentConfig *ApolloConnConfig
	currentJson,err:=json.Marshal(currentConnApolloConfig)
	test.Nil(t,err)

	t.Log("currentJson:",string(currentJson))

	test.Equal(t,false,string(currentJson)=="")

	json.Unmarshal(currentJson,&currentConfig)

	test.NotNil(t,currentConfig)

	updateApolloConfig(nil)

	//make sure currentConnApolloConfig was not modified
	test.Equal(t,currentConfig.NamespaceName,currentConnApolloConfig.NamespaceName)
	test.Equal(t,currentConfig.AppId,currentConnApolloConfig.AppId)
	test.Equal(t,currentConfig.Cluster,currentConnApolloConfig.Cluster)
	test.Equal(t,currentConfig.ReleaseKey,currentConnApolloConfig.ReleaseKey)

}

func TestGetConfigValueTimeout(t *testing.T) {
	expireTime:=5
	configs:=createMockApolloConfig(expireTime)

	for k,v:=range configs{
		test.Equal(t,v,getValue(k))
	}

	time.Sleep(time.Duration(expireTime)*time.Second)

	for k,_:=range configs{
		test.Equal(t,"",getValue(k))
	}
}

func TestGetConfigValueNullApolloConfig(t *testing.T) {
	//clear Configurations
	apolloConfigCache.Clear()

	//test getValue
	value:=getValue("joe")

	test.Equal(t,empty,value)

	//test GetStringValue
	defaultValue:="j"

	//test default
	v:=GetStringValue("joe",defaultValue)

	test.Equal(t,defaultValue,v)

	createMockApolloConfig(configCacheExpireTime)
}

func TestGetIntValue(t *testing.T) {
	defaultValue:=100000

	//test default
	v:=GetIntValue("joe",defaultValue)

	test.Equal(t,defaultValue,v)

	//normal value
	v=GetIntValue("int",defaultValue)

	test.Equal(t,1,v)

	//error type
	v=GetIntValue("float",defaultValue)

	test.Equal(t,defaultValue,v)
}

func TestGetFloatValue(t *testing.T) {
	defaultValue:=100000.1

	//test default
	v:=GetFloatValue("joe",defaultValue)

	test.Equal(t,defaultValue,v)

	//normal value
	v=GetFloatValue("float",defaultValue)

	test.Equal(t,190.3,v)

	//error type
	v=GetFloatValue("int",defaultValue)

	test.Equal(t,float64(1),v)
}

func TestGetBoolValue(t *testing.T) {
	defaultValue:=false

	//test default
	v:=GetBoolValue("joe",defaultValue)

	test.Equal(t,defaultValue,v)

	//normal value
	v=GetBoolValue("bool",defaultValue)

	test.Equal(t,true,v)

	//error type
	v=GetBoolValue("float",defaultValue)

	test.Equal(t,defaultValue,v)
}

func TestGetStringValue(t *testing.T) {
	defaultValue:="j"

	//test default
	v:=GetStringValue("joe",defaultValue)

	test.Equal(t,defaultValue,v)

	//normal value
	v=GetStringValue("string",defaultValue)

	test.Equal(t,"value",v)
}