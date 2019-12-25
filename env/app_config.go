package env

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/zouyx/agollo/v2/component"
	. "github.com/zouyx/agollo/v2/component/log"
	"github.com/zouyx/agollo/v2/component/notify"
	"github.com/zouyx/agollo/v2/utils"
)

const (
	APP_CONFIG_FILE_NAME = "app.properties"
	ENV_CONFIG_FILE_PATH = "AGOLLO_CONF"
)

func init() {
	//init config
	initFileConfig()
}

var (
	long_poll_connect_timeout = 1 * time.Minute //1m

	//for typed config agcache of parser result, e.g. integer, double, long, etc.
	//max_config_cache_size    = 500             //500 agcache key
	//config_cache_expire_time = 1 * time.Minute //1 minute

	//refresh ip list
	refresh_ip_list_interval = 20 * time.Minute //20m

	//appconfig
	appConfig *AppConfig

	//real servers ip
	servers sync.Map

	//next try connect period - 60 second
	next_try_connect_period int64 = 60
)

type AppConfig struct {
	AppId            string `json:"appId"`
	Cluster          string `json:"cluster"`
	NamespaceName    string `json:"namespaceName"`
	Ip               string `json:"ip"`
	NextTryConnTime  int64  `json:"-"`
	IsBackupConfig   bool   `default:"true" json:"isBackupConfig"`
	BackupConfigPath string `json:"backupConfigPath"`
}

//getIsBackupConfig whether backup config after fetch config from apollo
//false : no
//true : yes (default)
func (this *AppConfig) GetIsBackupConfig() bool {
	return this.IsBackupConfig
}

func (this *AppConfig) GetBackupConfigPath() string {
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

func (this *AppConfig) SelectHost() string {
	if !this.isConnectDirectly() {
		return this.getHost()
	}

	host := ""

	servers.Range(func(k, v interface{}) bool {
		server := v.(*serverInfo)
		// if some node has down then select next node
		if server.IsDown {
			return true
		}
		host = k.(string)
		return false
	})

	return host
}

func SetDownNode(host string) {
	if host == "" || appConfig == nil {
		return
	}

	if host == appConfig.getHost() {
		appConfig.setNextTryConnTime(next_try_connect_period)
	}

	servers.Range(func(k, v interface{}) bool {
		server := v.(*serverInfo)
		// if some node has down then select next node
		if k.(string) == host {
			server.IsDown = true
			return false
		}
		return true
	})
}

type serverInfo struct {
	AppName     string `json:"appName"`
	InstanceId  string `json:"instanceId"`
	HomepageUrl string `json:"homepageUrl"`
	IsDown      bool   `json:"-"`
}

func initFileConfig() {
	// default use application.properties
	InitConfig(nil)
}

func InitConfig(loadAppConfig func() (*AppConfig, error)) {
	var err error
	//init config file
	appConfig, err = getLoadAppConfig(loadAppConfig)

	if err != nil {
		return
	}
	initApolloConfigCache(appConfig.NamespaceName)
}

//initApolloConfigCache 根据namespace初始化apollo配置
func initApolloConfigCache(namespace string) {
	func(appConfig *AppConfig) {
		notify.SplitNamespaces(namespace, func(namespace string) {
			apolloConfig := &component.ApolloConfig{}
			apolloConfig.Init(
				appConfig.AppId,
				appConfig.Cluster,
				namespace)

			go updateApolloConfig(apolloConfig, false)
		})
	}(appConfig)
}

// set load app config's function
func getLoadAppConfig(loadAppConfig func() (*AppConfig, error)) (*AppConfig, error) {
	if loadAppConfig != nil {
		return loadAppConfig()
	}
	configPath := os.Getenv(ENV_CONFIG_FILE_PATH)
	if configPath == "" {
		configPath = APP_CONFIG_FILE_NAME
	}
	return loadJsonConfig(configPath)
}

//set timer for update ip list
//interval : 20m
func initServerIpList() {
	syncServerIpList(nil)
	Logger.Debug("syncServerIpList started")

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
	Logger.Debug("get all server info:", string(responseBody))

	tmpServerInfo := make([]*serverInfo, 0)

	err = json.Unmarshal(responseBody, &tmpServerInfo)

	if err != nil {
		Logger.Error("Unmarshal json Fail,Error:", err)
		return
	}

	if len(tmpServerInfo) == 0 {
		Logger.Info("get no real server!")
		return
	}

	for _, server := range tmpServerInfo {
		if server == nil {
			continue
		}
		servers.Store(server.HomepageUrl, server)
	}
	return
}

//sync ip list from server
//then
//1.update agcache
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
	return fmt.Sprintf("%sconfigs/%s/%s/%s?releaseKey=%s&ip=%s",
		host,
		url.QueryEscape(config.AppId),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(config.NamespaceName),
		url.QueryEscape(component.GetCurrentApolloConfigReleaseKey(config.NamespaceName)),
		utils.GetInternal())
}

func getServicesConfigUrl(config *AppConfig) string {
	return fmt.Sprintf("%sservices/config?appId=%s&ip=%s",
		config.getHost(),
		url.QueryEscape(config.AppId),
		utils.GetInternal())
}

func GetPlainAppConfig() *AppConfig {
	return appConfig
}
