package agollo

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const appConfigFileName = "app.properties"

var (
	long_poll_interval        = 2 * time.Second //2s
	long_poll_connect_timeout = 1 * time.Minute //1m

	connect_timeout = 1 * time.Second //1s
	//notify timeout
	nofity_connect_timeout = 10 * time.Minute //10m
	//for on error retry
	on_error_retry_interval = 1 * time.Second //1s
	//for typed config cache of parser result, e.g. integer, double, long, etc.
	//max_config_cache_size    = 500             //500 cache key
	//config_cache_expire_time = 1 * time.Minute //1 minute

	//max retries connect apollo
	max_retries = 5

	//refresh ip list
	refresh_ip_list_interval = 20 * time.Minute //20m

	//appconfig
	appConfig *AppConfig

	//real servers ip
	servers map[string]*serverInfo = make(map[string]*serverInfo, 0)

	//next try connect period - 60 second
	next_try_connect_period int64 = 60
)

type AppConfig struct {
	AppId            string `json:"appId"`
	Cluster          string `json:"cluster"`
	NamespaceName    string `json:"namespaceName"`
	Ip               string `json:"ip"`
	NextTryConnTime  int64  `json:"-"`
	BackupConfigPath string `json:"backupConfigPath"`
}

func (this *AppConfig) getBackupConfigPath() string {
	return this.BackupConfigPath
}

func (this *AppConfig) getHost() string {
	if strings.HasPrefix(this.Ip, "http") {
		if !strings.HasSuffix(this.Ip, "/") {
			return this.Ip + "/"
		}
		return this.Ip
	}
	return "http://" + this.Ip + "/"
}

//if this connect is fail will set this time
func (this *AppConfig) setNextTryConnTime(nextTryConnectPeriod int64) {
	this.NextTryConnTime = time.Now().Unix() + nextTryConnectPeriod
}

//is connect by ip directly
//false : no
//true : yes
func (this *AppConfig) isConnectDirectly() bool {
	if this.NextTryConnTime >= 0 && this.NextTryConnTime > time.Now().Unix() {
		return true
	}

	return false
}

func (this *AppConfig) selectHost() string {
	if !this.isConnectDirectly() {
		return this.getHost()
	}

	for host, server := range servers {
		// if some node has down then select next node
		if server.IsDown {
			continue
		}
		return host
	}

	return ""
}

func setDownNode(host string) {
	if host == "" || appConfig == nil {
		return
	}

	if host == appConfig.getHost() {
		appConfig.setNextTryConnTime(next_try_connect_period)
	}

	for key, server := range servers {
		if key == host {
			server.IsDown = true
			break
		}
	}
}

type serverInfo struct {
	AppName     string `json:"appName"`
	InstanceId  string `json:"instanceId"`
	HomepageUrl string `json:"homepageUrl"`
	IsDown      bool   `json:"-"`
}

func init() {
	//init config
	initFileConfig()

	//init common
	initCommon()
}

func initCommon() {
	//init server ip list
	go initServerIpList()
}

func initFileConfig() {
	// default use application.properties
	initConfig(nil)
}

func initConfig(loadAppConfig func() (*AppConfig, error)) {
	var err error
	//init config file
	appConfig, err = getLoadAppConfig(loadAppConfig)

	if err != nil {
		//增加当配置文件不存在时，从环境变量中读取配置
		appId := filepath.Base(os.Args[0]) //获取当前执行文件名字
		appId = strings.Replace(appId, "-linux-amd64", "", 1)
		aConfig := &AppConfig{
			Cluster:       default_cluster,
			NamespaceName: default_namespace,
			AppId:         appId,
			Ip:            os.Getenv("APOLLO_META"),
		}
		appConfig = aConfig
		err = nil
		return
	}

	func(appConfig *AppConfig) {
		apolloConfig := &ApolloConfig{}
		apolloConfig.AppId = appConfig.AppId
		apolloConfig.Cluster = appConfig.Cluster
		apolloConfig.NamespaceName = appConfig.NamespaceName

		updateApolloConfig(apolloConfig, false)
	}(appConfig)
}

//init config by custom
func InitCustomConfig(loadAppConfig func() (*AppConfig, error)) {

	initConfig(loadAppConfig)

	//init all notification
	initAllNotifications()

}

// set load app config's function
func getLoadAppConfig(loadAppConfig func() (*AppConfig, error)) (*AppConfig, error) {
	if loadAppConfig != nil {
		return loadAppConfig()
	}
	return loadJsonConfig(appConfigFileName)
}

//set timer for update ip list
//interval : 20m
func initServerIpList() {
	syncServerIpList(nil)

	t2 := time.NewTimer(refresh_ip_list_interval)
	for {
		select {
		case <-t2.C:
			syncServerIpList(nil)
			t2.Reset(refresh_ip_list_interval)
		}
	}
}

func syncServerIpListSuccessCallBack(responseBody []byte) (o interface{}, err error) {
	logger.Debug("get all server info:", string(responseBody))

	tmpServerInfo := make([]*serverInfo, 0)

	err = json.Unmarshal(responseBody, &tmpServerInfo)

	if err != nil {
		logger.Error("Unmarshal json Fail,Error:", err)
		return
	}

	if len(tmpServerInfo) == 0 {
		logger.Info("get no real server!")
		return
	}

	for _, server := range tmpServerInfo {
		if server == nil {
			continue
		}
		servers[server.HomepageUrl] = server
	}
	return
}

//sync ip list from server
//then
//1.update cache
//2.store in disk
func syncServerIpList(newAppConfig *AppConfig) error {
	appConfig := GetAppConfig(newAppConfig)
	if appConfig == nil {
		panic("can not find apollo config!please confirm!")
	}

	_, err := request(getServicesConfigUrl(appConfig), &ConnectConfig{}, &CallBack{
		SuccessCallBack: syncServerIpListSuccessCallBack,
	})

	return err
}

func GetAppConfig(newAppConfig *AppConfig) *AppConfig {
	if newAppConfig != nil {
		return newAppConfig
	}
	return appConfig
}

func getConfigUrl(config *AppConfig) string {
	return getConfigUrlByHost(config, config.getHost())
}

func getConfigUrlByHost(config *AppConfig, host string) string {
	current := GetCurrentApolloConfig()
	return fmt.Sprintf("%sconfigs/%s/%s/%s?releaseKey=%s&ip=%s",
		host,
		url.QueryEscape(config.AppId),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(config.NamespaceName),
		url.QueryEscape(current.ReleaseKey),
		getInternal())
}

func getConfigUrlSuffix(config *AppConfig, newConfig *AppConfig) string {
	if newConfig != nil {
		return ""
	}
	current := GetCurrentApolloConfig()
	return fmt.Sprintf("configs/%s/%s/%s?releaseKey=%s&ip=%s",
		url.QueryEscape(config.AppId),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(config.NamespaceName),
		url.QueryEscape(current.ReleaseKey),
		getInternal())
}

func getNotifyUrlSuffix(notifications string, config *AppConfig, newConfig *AppConfig) string {
	if newConfig != nil {
		return ""
	}
	return fmt.Sprintf("notifications/v2?appId=%s&cluster=%s&notifications=%s",
		url.QueryEscape(config.AppId),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(notifications))
}

func getServicesConfigUrl(config *AppConfig) string {
	return fmt.Sprintf("%sservices/config?appId=%s&ip=%s",
		config.getHost(),
		url.QueryEscape(config.AppId),
		getInternal())
}
