package agollo

import (
	"github.com/zouyx/agollo/test"
	"testing"
	"time"
)

func TestRequestRecovery(t *testing.T) {
	time.Sleep(1 * time.Second)
	mockIpList(t)
	server := runNormalBackupConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	appConfig := GetAppConfig(newAppConfig)
	urlSuffix := getConfigUrlSuffix(appConfig, newAppConfig)

	o, err := requestRecovery(appConfig, &ConnectConfig{
		Uri: urlSuffix,
	}, &CallBack{
		SuccessCallBack: autoSyncConfigServicesSuccessCallBack,
	})

	test.Nil(t, err)
	test.Nil(t, o)
}

func TestCustomTimeout(t *testing.T) {
	time.Sleep(1 * time.Second)
	mockIpList(t)
	server := runLongTimeResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	startTime := time.Now().Second()
	appConfig := GetAppConfig(newAppConfig)
	urlSuffix := getConfigUrlSuffix(appConfig, newAppConfig)

	o, err := requestRecovery(appConfig, &ConnectConfig{
		Uri:     urlSuffix,
		Timeout: 11 * time.Second,
	}, &CallBack{
		SuccessCallBack: autoSyncConfigServicesSuccessCallBack,
	})

	endTime := time.Now().Second()
	t.Log("starttime:", startTime)
	t.Log("endTime:", endTime)
	t.Log("duration:", endTime-startTime)
	test.Equal(t, 10, endTime-startTime)
	test.Nil(t, err)
	test.Nil(t, o)
}

func mockIpList(t *testing.T) {
	server := runNormalServicesResponse()
	defer server.Close()
	time.Sleep(1 * time.Second)
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	err := syncServerIpList(newAppConfig)

	test.Nil(t, err)

	test.Equal(t, 2, len(servers))
}
