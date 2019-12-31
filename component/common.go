package component

import (
	"fmt"
	"github.com/zouyx/agollo/v2/env/config"
	"net/url"
	"time"

	. "github.com/zouyx/agollo/v2/component/log"
	"github.com/zouyx/agollo/v2/protocol/http"

	"github.com/zouyx/agollo/v2/env"
	"github.com/zouyx/agollo/v2/utils"
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

type AbsComponent interface {
	Start()
}

func StartRefreshConfig(component AbsComponent) {
	component.Start()
}

func GetConfigURLSuffix(config *config.AppConfig, namespaceName string) string {
	if config == nil {
		return ""
	}
	return fmt.Sprintf("configs/%s/%s/%s?releaseKey=%s&ip=%s",
		url.QueryEscape(config.AppId),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(namespaceName),
		url.QueryEscape(env.GetCurrentApolloConfigReleaseKey(namespaceName)),
		utils.GetInternal())
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
