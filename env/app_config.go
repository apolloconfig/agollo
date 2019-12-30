package env

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	. "github.com/zouyx/agollo/v2/component/log"
	"github.com/zouyx/agollo/v2/utils"
)

const (
	APP_CONFIG_FILE_NAME = "../app.properties"
	ENV_CONFIG_FILE_PATH = "AGOLLO_CONF"

	default_notification_id = -1
	comma                   = ","
)

func init() {
	//init config
	InitFileConfig()
}

var (
	long_poll_connect_timeout = 1 * time.Minute //1m

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

func (this *AppConfig) GetHost() string {
	if strings.HasPrefix(this.Ip, "http") {
		if !strings.HasSuffix(this.Ip, "/") {
			return this.Ip + "/"
		}
		return this.Ip
	}
	return "http://" + this.Ip + "/"
}

//if this connect is fail will set this time
func (this *AppConfig) SetNextTryConnTime(nextTryConnectPeriod int64) {
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
		return this.GetHost()
	}

	host := ""

	GetServers().Range(func(k, v interface{}) bool {
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

	if host == appConfig.GetHost() {
		appConfig.SetNextTryConnTime(next_try_connect_period)
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

func InitFileConfig() {
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
}

func SplitNamespaces(namespacesStr string, callback func(namespace string)) map[string]int64 {
	namespaces := make(map[string]int64, 1)
	split := strings.Split(namespacesStr, comma)
	for _, namespace := range split {
		callback(namespace)
		namespaces[namespace] = default_notification_id
	}
	return namespaces
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

func SyncServerIpListSuccessCallBack(responseBody []byte) (o interface{}, err error) {
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

func GetAppConfig(newAppConfig *AppConfig) *AppConfig {
	if newAppConfig != nil {
		return newAppConfig
	}
	return appConfig
}

func GetServicesConfigUrl(config *AppConfig) string {
	return fmt.Sprintf("%sservices/config?appId=%s&ip=%s",
		config.GetHost(),
		url.QueryEscape(config.AppId),
		utils.GetInternal())
}

func GetPlainAppConfig() *AppConfig {
	return appConfig
}

func GetServers() *sync.Map {
	return &servers
}

func GetServersLen()int  {
	s := GetServers()
	l := 0
	s.Range(func(k, v interface{}) bool {
		l++
		return true
	})
	return l
}