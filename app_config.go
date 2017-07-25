package agollo

import (
	"encoding/json"
	"os"
	"strconv"
	"github.com/cihub/seelog"
	"time"
	"fmt"
	"net/url"
)

var (
	refresh_interval = 5 *time.Minute //5m
	refresh_interval_key = "apollo.refreshInterval"  //

	long_poll_interval = 5 *time.Second //5s
	long_poll_connect_timeout  = 1 * time.Minute //1m

	connect_timeout  = 1 * time.Second //1s
	read_timeout     = 5 * time.Second //5s
	//for on error retry
	ON_ERROR_RETRY_INTERVAL = 1 * time.Second //1s
	//for typed config cache of parser result, e.g. integer, double, long, etc.
	max_config_cache_size    = 500             //500 cache key
	config_cache_expire_time = 1 * time.Minute //1 minute

	//max retries connect apollo
	max_retries=5

	//appconfig
	appConfig *AppConfig
)

func init() {
	//init common
	initRefreshInterval()
}

func init() {
	//init config file
	appConfig = LoadJsonConfig()

	go func(appConfig *AppConfig) {
		apolloConfig:=&ApolloConfig{
			AppId:appConfig.AppId,
			Cluster:appConfig.Cluster,
			NamespaceName:appConfig.NamespaceName,
		}

		UpdateApolloConfig(apolloConfig)
	}(appConfig)


}

func GetAppConfig()*AppConfig  {
	return appConfig
}

type AppConfig struct {
	AppId string `json:"appId"`
	Cluster string `json:"cluster"`
	NamespaceName string `json:"namespaceName"`
	ReleaseKey string `json:"releaseKey"`
	Ip string `json:"ip"`
}

func CreateAppConfigWithJson(str string) (*AppConfig,error) {
	appConfig:=&AppConfig{}
	err:=json.Unmarshal([]byte(str),appConfig)
	if IsNotNil(err) {
		return nil,err
	}
	return appConfig,nil
}

func initRefreshInterval() error {
	customizedRefreshInterval:=os.Getenv(refresh_interval_key)
	if IsNotEmpty(customizedRefreshInterval){
		interval,err:=strconv.Atoi(customizedRefreshInterval)
		if IsNotNil(err) {
			seelog.Errorf("Config for apollo.refreshInterval is invalid:%s",customizedRefreshInterval)
			return err
		}
		refresh_interval=time.Duration(interval)
	}
	return nil
}

func GetConfigUrl(config *AppConfig) string{
	current:=GetCurrentApolloConfig()
	return fmt.Sprintf("http://%s/configs/%s/%s/%s?releaseKey=%s&ip=%s",
		config.Ip,
		url.QueryEscape(current.AppId),
		url.QueryEscape(current.Cluster),
		url.QueryEscape(current.NamespaceName),
		url.QueryEscape(current.ReleaseKey),
		GetInternal())
}

func GetNotifyUrl(notifications string,config *AppConfig) string{
	current:=GetCurrentApolloConfig()
	return fmt.Sprintf("http://%s/notifications/v2?appId=%s&cluster=%s&notifications=%s",
		config.Ip,
		url.QueryEscape(current.AppId),
		url.QueryEscape(current.Cluster),
		url.QueryEscape(notifications))
}