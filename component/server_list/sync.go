package server_list

import (
	. "github.com/zouyx/agollo/v2/component/log"
	"github.com/zouyx/agollo/v2/env"
	"github.com/zouyx/agollo/v2/env/config"
	"github.com/zouyx/agollo/v2/protocol/http"
	"time"
)

const (
	//refresh ip list
	refresh_ip_list_interval = 20 * time.Minute //20m
)

func init() {
	go InitServerIpList()
}

//set timer for update ip list
//interval : 20m
func InitServerIpList() {
	SyncServerIpList(nil)
	Logger.Debug("syncServerIpList started")

	t2 := time.NewTimer(refresh_ip_list_interval)
	for {
		select {
		case <-t2.C:
			SyncServerIpList(nil)
			t2.Reset(refresh_ip_list_interval)
		}
	}
}

//sync ip list from server
//then
//1.update agcache
//2.store in disk
func SyncServerIpList(newAppConfig *config.AppConfig) error {
	appConfig := env.GetAppConfig(newAppConfig)
	if appConfig == nil {
		panic("can not find apollo config!please confirm!")
	}

	_, err := http.Request(env.GetServicesConfigUrl(appConfig), &env.ConnectConfig{}, &http.CallBack{
		SuccessCallBack: env.SyncServerIpListSuccessCallBack,
	})

	return err
}
