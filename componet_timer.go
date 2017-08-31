package agollo

import (
	"time"
	"github.com/cihub/seelog"
)

type AutoRefreshConfigComponent struct {

}

func (this *AutoRefreshConfigComponent) Start()  {
	for {
			notifySyncConfigServices()
			time.Sleep(refresh_interval)
		}
}

func SyncConfig() error {
	return autoSyncConfigServices()
}


func autoSyncConfigServicesSuccessCallBack(responseBody []byte)(o interface{},err error){
	apolloConfig,err:=createApolloConfigWithJson(responseBody)

	if err!=nil{
		seelog.Error("Unmarshal Msg Fail,Error:",err)
		return nil,err
	}

	updateApolloConfig(apolloConfig)

	return nil,nil
}

func autoSyncConfigServices() error {
	appConfig:=GetAppConfig()
	if appConfig==nil{
		panic("can not find apollo config!please confirm!")
	}


	urlSuffix:=getConfigUrlSuffix(appConfig)

	_,err:=requestRecovery(appConfig,urlSuffix,autoSyncConfigServicesSuccessCallBack)

	return err
}
