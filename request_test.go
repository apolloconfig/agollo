package agollo

import (
	"testing"
	"time"
	"github.com/zouyx/agollo/test"
)

func TestRequestRecovery(t *testing.T) {
	time.Sleep(1*time.Second)
	mockIpList(t)
	go runMockConfigBackupServer(normalBackupConfigResponse)
	defer closeAllMockServicesServer()

	appConfig:=GetAppConfig()
	urlSuffix:=getConfigUrlSuffix(appConfig)

	o,err:=requestRecovery(appConfig,&ConnectConfig{
		Uri:urlSuffix,
	},&CallBack{
		SuccessCallBack:autoSyncConfigServicesSuccessCallBack,
	})

	test.Nil(t,err)
	test.Nil(t,o)
}

func TestCustomTimeout(t *testing.T) {
	time.Sleep(1*time.Second)
	mockIpList(t)
	go runMockConfigBackupServer(longTimeResponse)
	defer closeAllMockServicesServer()

	startTime := time.Now().Second()
	appConfig:=GetAppConfig()
	urlSuffix:=getConfigUrlSuffix(appConfig)

	o,err:=requestRecovery(appConfig,&ConnectConfig{
		Uri:urlSuffix,
		Timeout:11*time.Second,
	},&CallBack{
		SuccessCallBack:autoSyncConfigServicesSuccessCallBack,
	})

	endTime := time.Now().Second()
	t.Log("starttime:",startTime)
	t.Log("endTime:",endTime)
	t.Log("duration:",endTime-startTime)
	test.Equal(t,10,endTime-startTime)
	test.Nil(t,err)
	test.Nil(t,o)
}

//func TestErrorRequestRecovery(t *testing.T) {
//	time.Sleep(1*time.Second)
//	mockBackupConfig()
//	go runMockConfigBackupServer(errorBackupConfigResponse)
//	defer closeAllMockServicesServer()
//
//	appConfig:=GetAppConfig()
//	urlSuffix:=getConfigUrlSuffix(appConfig)
//
//	o,err:=requestRecovery(appConfig,urlSuffix,autoSyncConfigServicesSuccessCallBack)
//
//	test.NotNil(t,err)
//	test.Nil(t,o)
//}
//
//func mockBackupConfig(){
//	syncServerIpListSuccessCallBack([]byte(servicesResponseStr))
//}

func mockIpList(t *testing.T) {
	go runMockServicesServer(normalServicesResponse)
	defer closeMockServicesServer()
	time.Sleep(1*time.Second)

	err:=syncServerIpList()

	test.Nil(t,err)

	test.Equal(t,2,len(servers))
}