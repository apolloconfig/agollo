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
	REFRESH_INTERVAL = 5 *time.Minute //5m
	REFRESH_INTERVAL_KEY = "apollo.refreshInterval"  //

	LONG_POLL_INTERVAL = 5 *time.Second //5s
	LONG_POLL_CONNECT_TIMEOUT  = 1 * time.Minute //1m

	CONNECT_TIMEOUT  = 1 * time.Second //1s
	READ_TIMEOUT     = 5 * time.Second //5s
	LOAD_CONFIG_QPS  = 2
	LONG_POLL_QPS    = 2
	//for on error retry
	ON_ERROR_RETRY_INTERVAL = 1 * time.Second //1s
	//for typed config cache of parser result, e.g. integer, double, long, etc.
	MAX_CONFIG_CACHE_SIZE    = 500             //500 cache key
	CONFIG_CACHE_EXPIRE_TIME = 1 * time.Minute //1 minute

	//max retries connect apollo
	MAX_RETRIES=5

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
	customizedRefreshInterval:=os.Getenv(REFRESH_INTERVAL_KEY)
	if IsNotEmpty(customizedRefreshInterval){
		interval,err:=strconv.Atoi(customizedRefreshInterval)
		if IsNotNil(err) {
			seelog.Errorf("Config for apollo.refreshInterval is invalid:%s",customizedRefreshInterval)
			return err
		}
		REFRESH_INTERVAL=time.Duration(interval)
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