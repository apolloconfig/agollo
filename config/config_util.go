package config

import (
	"time"
	"os"
	"strconv"
	"github.com/zouyx/agollo/utils/stringutils"
	"github.com/zouyx/agollo/utils/objectutils"

	"github.com/cihub/seelog"
	_ "github.com/zouyx/agollo/utils/logs"
	"fmt"
	"net/url"
	"github.com/zouyx/agollo/utils"
	"github.com/zouyx/agollo/repository"
	"github.com/zouyx/agollo/dto"
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
)

func init() {
	//init common
	initRefreshInterval()
}

func initRefreshInterval() error {
	customizedRefreshInterval:=os.Getenv(REFRESH_INTERVAL_KEY)
	if stringutils.IsNotEmpty(customizedRefreshInterval){
		interval,err:=strconv.Atoi(customizedRefreshInterval)
		if objectutils.IsNotNil(err) {
			seelog.Errorf("Config for apollo.refreshInterval is invalid:%s",customizedRefreshInterval)
			return err
		}
		REFRESH_INTERVAL=time.Duration(interval)
	}
	return nil
}

func GetConfigUrl(config *dto.AppConfig) string{
	current:=repository.GetCurrentApolloConfig()
	return fmt.Sprintf("http://%s/configs/%s/%s/%s?releaseKey=%s&ip=%s",
		config.Ip,
		url.QueryEscape(current.AppId),
		url.QueryEscape(current.Cluster),
		url.QueryEscape(current.NamespaceName),
		url.QueryEscape(current.ReleaseKey),
		utils.GetInternal())
}

func GetNotifyUrl(notifications string,config *dto.AppConfig) string{
	current:=repository.GetCurrentApolloConfig()
	return fmt.Sprintf("http://%s/notifications/v2?appId=%s&cluster=%s&notifications=%s",
		config.Ip,
		url.QueryEscape(current.AppId),
		url.QueryEscape(current.Cluster),
		url.QueryEscape(notifications))
}