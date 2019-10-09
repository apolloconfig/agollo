package agollo

import (
	. "github.com/tevid/gohamcrest"
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
	urlSuffix := getConfigUrlSuffix(appConfig, newAppConfig.NamespaceName)

	o, err := requestRecovery(appConfig, &ConnectConfig{
		Uri: urlSuffix,
	}, &CallBack{
		SuccessCallBack: autoSyncConfigServicesSuccessCallBack,
	})

	Assert(t, err,NilVal())
	Assert(t, o,NilVal())
}

func TestCustomTimeout(t *testing.T) {
	time.Sleep(1 * time.Second)
	mockIpList(t)
	server := runLongTimeResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	startTime := time.Now().Second()
	appConfig := GetAppConfig(newAppConfig)
	urlSuffix := getConfigUrlSuffix(appConfig, newAppConfig.NamespaceName)

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
	Assert(t, 10, Equal(endTime-startTime))
	Assert(t, err,NilVal())
	Assert(t, o,NilVal())
}

func mockIpList(t *testing.T) {
	server := runNormalServicesResponse()
	defer server.Close()
	time.Sleep(1 * time.Second)
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	err := syncServerIpList(newAppConfig)

	Assert(t, err,NilVal())

	Assert(t, 2, Equal(len(servers)))
}
