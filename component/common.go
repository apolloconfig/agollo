package component

import (
	. "github.com/zouyx/agollo/v2/component/log"
	"fmt"
	"github.com/zouyx/agollo/v2/protocol/http"
	"net/url"
	"sync"
	"time"

	"github.com/zouyx/agollo/v2/env"
	"github.com/zouyx/agollo/v2/utils"
)

const (
	//refresh ip list
	refresh_ip_list_interval = 20 * time.Minute //20m
)

var (
	currentConnApolloConfig = &currentApolloConfig{
		configs: make(map[string]*env.ApolloConnConfig, 1),
	}
)

func init() {
	InitServerIpList()
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

type currentApolloConfig struct {
	l       sync.RWMutex
	configs map[string]*env.ApolloConnConfig
}

func SetCurrentApolloConfig(namespace string, connConfig *env.ApolloConnConfig) {
	currentConnApolloConfig.l.Lock()
	defer currentConnApolloConfig.l.Unlock()

	currentConnApolloConfig.configs[namespace] = connConfig
}

//GetCurrentApolloConfig 获取Apollo链接配置
func GetCurrentApolloConfig() map[string]*env.ApolloConnConfig {
	currentConnApolloConfig.l.RLock()
	defer currentConnApolloConfig.l.RUnlock()

	return currentConnApolloConfig.configs
}

func GetCurrentApolloConfigReleaseKey(namespace string) string {
	currentConnApolloConfig.l.RLock()
	defer currentConnApolloConfig.l.RUnlock()
	config := currentConnApolloConfig.configs[namespace]
	if config == nil {
		return utils.Empty
	}

	return config.ReleaseKey
}

func GetConfigURLSuffix(config *env.AppConfig, namespaceName string) string {
	if config == nil {
		return ""
	}
	return fmt.Sprintf("configs/%s/%s/%s?releaseKey=%s&ip=%s",
		url.QueryEscape(config.AppId),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(namespaceName),
		url.QueryEscape(GetCurrentApolloConfigReleaseKey(namespaceName)),
		utils.GetInternal())
}


//sync ip list from server
//then
//1.update agcache
//2.store in disk
func SyncServerIpList(newAppConfig *env.AppConfig) error {
	appConfig := env.GetAppConfig(newAppConfig)
	if appConfig == nil {
		panic("can not find apollo config!please confirm!")
	}

	_, err := http.Request(env.GetServicesConfigUrl(appConfig), &env.ConnectConfig{}, &http.CallBack{
		SuccessCallBack: env.SyncServerIpListSuccessCallBack,
	})

	return err
}