package agollo

import (
	"time"
)

type AutoRefreshConfigComponent struct {

}

func (this *AutoRefreshConfigComponent) Start()  {
	t2 := time.NewTimer(refresh_interval)
	for {
		select {
		case <-t2.C:
			notifySyncConfigServices()
			t2.Reset(refresh_interval)
		}
	}
}

func SyncConfig() error {
	return autoSyncConfigServices(nil)
}


func autoSyncConfigServicesSuccessCallBack(responseBody []byte)(o interface{},err error){
	apolloConfig,err:=createApolloConfigWithJson(responseBody)

	if err!=nil{
		logger.Error("Unmarshal Msg Fail,Error:",err)
		return nil,err
	}

	updateApolloConfig(apolloConfig,true)

	return nil,nil
}

func autoSyncConfigServices(newAppConfig *AppConfig) error {
	appConfig:=GetAppConfig(newAppConfig)
	if appConfig==nil{
		panic("can not find apollo config!please confirm!")
	}

	urlSuffix:=getConfigUrlSuffix(appConfig,newAppConfig)

	_,err:=requestRecovery(appConfig,&ConnectConfig{
		Uri:urlSuffix,
	},&CallBack{
		SuccessCallBack:autoSyncConfigServicesSuccessCallBack,
		NotModifyCallBack:touchApolloConfigCache,
	})

	return err
}
