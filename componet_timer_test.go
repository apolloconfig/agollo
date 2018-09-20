package agollo

import (
	"testing"
	"github.com/zouyx/agollo/test"
	"time"
	"fmt"
)

func TestAutoSyncConfigServices(t *testing.T) {
	server := runNormalConfigResponse()
	newAppConfig:=getTestAppConfig()
	newAppConfig.Ip=server.URL

	time.Sleep(1*time.Second)

	appConfig.NextTryConnTime=0

	err:=autoSyncConfigServices(newAppConfig)
	err=autoSyncConfigServices(newAppConfig)

	test.Nil(t,err)

	config:=GetCurrentApolloConfig()

	test.Equal(t,"100004458",config.AppId)
	test.Equal(t,"default",config.Cluster)
	test.Equal(t,"application",config.NamespaceName)
	test.Equal(t,"20170430092936-dee2d58e74515ff3",config.ReleaseKey)
	//test.Equal(t,"value1",config.Configurations["key1"])
	//test.Equal(t,"value2",config.Configurations["key2"])
}

func TestAutoSyncConfigServicesNormal2NotModified(t *testing.T) {
	server := runLongNotmodifiedConfigResponse()
	newAppConfig:=getTestAppConfig()
	newAppConfig.Ip=server.URL
	time.Sleep(1*time.Second)

	appConfig.NextTryConnTime=0

	autoSyncConfigServicesSuccessCallBack([]byte(configResponseStr))

	config:=GetCurrentApolloConfig()

	fmt.Println("sleeping 10s")

	time.Sleep(10*time.Second)

	fmt.Println("checking cache time left")
	it := apolloConfigCache.NewIterator()
	for i := int64(0); i < apolloConfigCache.EntryCount(); i++ {
		entry := it.Next()
		if entry==nil{
			break
		}
		timeLeft, err := apolloConfigCache.TTL([]byte(entry.Key))
		test.Nil(t,err)
		fmt.Printf("key:%s,time:%v \n",string(entry.Key),timeLeft)
		test.Equal(t,timeLeft>=110,true)
	}

	test.Equal(t,"100004458",config.AppId)
	test.Equal(t,"default",config.Cluster)
	test.Equal(t,"application",config.NamespaceName)
	test.Equal(t,"20170430092936-dee2d58e74515ff3",config.ReleaseKey)
	test.Equal(t,"value1",getValue("key1"))
	test.Equal(t,"value2",getValue("key2"))

	err:=autoSyncConfigServices(newAppConfig)

	fmt.Println("checking cache time left")
	it1 := apolloConfigCache.NewIterator()
	for i := int64(0); i < apolloConfigCache.EntryCount(); i++ {
		entry := it1.Next()
		if entry==nil{
			break
		}
		timeLeft, err := apolloConfigCache.TTL([]byte(entry.Key))
		test.Nil(t,err)
		fmt.Printf("key:%s,time:%v \n",string(entry.Key),timeLeft)
		test.Equal(t,timeLeft>=120,true)
	}

	fmt.Println(err)

	//sleep for async
	time.Sleep(1 *time.Second)
	checkBackupFile(t)
}

func checkBackupFile(t *testing.T){
	newConfig,e := loadConfigFile(appConfig.getBackupConfigPath())
	t.Log(newConfig.Configurations)
	isNil(e)
	isNotNil(newConfig.Configurations)
	for k,v :=range newConfig.Configurations  {
		test.Equal(t,getValue(k),v)
	}
}


//test if not modify
func TestAutoSyncConfigServicesNotModify(t *testing.T) {
	server := runNotModifyConfigResponse()
	newAppConfig:=getTestAppConfig()
	newAppConfig.Ip=server.URL

	apolloConfig,err:=createApolloConfigWithJson([]byte(configResponseStr))
	updateApolloConfig(apolloConfig,true)

	time.Sleep(10*time.Second)
	checkCacheLeft(t,configCacheExpireTime-10)

	appConfig.NextTryConnTime=0

	err=autoSyncConfigServices(newAppConfig)

	test.Nil(t,err)

	config:=GetCurrentApolloConfig()

	test.Equal(t,"100004458",config.AppId)
	test.Equal(t,"default",config.Cluster)
	test.Equal(t,"application",config.NamespaceName)
	test.Equal(t,"20170430092936-dee2d58e74515ff3",config.ReleaseKey)

	checkCacheLeft(t,configCacheExpireTime)

	//test.Equal(t,"value1",config.Configurations["key1"])
	//test.Equal(t,"value2",config.Configurations["key2"])
}



func TestAutoSyncConfigServicesError(t *testing.T) {
	//reload app properties
	go initFileConfig()
	server := runErrorConfigResponse()
	newAppConfig:=getTestAppConfig()
	newAppConfig.Ip=server.URL

	time.Sleep(1*time.Second)

	err:=autoSyncConfigServices(nil)

	test.NotNil(t,err)

	config:=GetCurrentApolloConfig()

	//still properties config
	test.Equal(t,"test",config.AppId)
	test.Equal(t,"dev",config.Cluster)
	test.Equal(t,"application",config.NamespaceName)
	test.Equal(t,"",config.ReleaseKey)
}