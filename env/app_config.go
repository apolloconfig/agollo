package env

import (
	"encoding/json"
	"fmt"
	"github.com/zouyx/agollo/v2/env/config"
	"github.com/zouyx/agollo/v2/env/config/json_config"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	. "github.com/zouyx/agollo/v2/component/log"
	"github.com/zouyx/agollo/v2/utils"
)

const (
	APP_CONFIG_FILE_NAME = "app.properties"
	ENV_CONFIG_FILE_PATH = "AGOLLO_CONF"

	default_notification_id = -1
	comma                   = ","

	default_cluster   = "default"
	default_namespace = "application"
)

var (
	//appconfig
	appConfig *config.AppConfig
	//real servers ip
	servers sync.Map

	long_poll_connect_timeout = 1 * time.Minute //1m

	//next try connect period - 60 second
	next_try_connect_period int64 = 60
)

func init() {
	//init config
	InitFileConfig()
}

func InitFileConfig() {
	// default use application.properties
	InitConfig(nil)
}

func InitConfig(loadAppConfig func() (*config.AppConfig, error)) {
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
		if callback != nil {
			callback(namespace)
		}
		namespaces[namespace] = default_notification_id
	}
	return namespaces
}

// set load app config's function
func getLoadAppConfig(loadAppConfig func() (*config.AppConfig, error)) (*config.AppConfig, error) {
	if loadAppConfig != nil {
		return loadAppConfig()
	}
	configPath := os.Getenv(ENV_CONFIG_FILE_PATH)
	if configPath == "" {
		configPath = APP_CONFIG_FILE_NAME
	}
	c, e := GetExecuteGetConfigFile().Load(configPath, Unmarshal)
	if c == nil {
		return nil, e
	}

	return c.(*config.AppConfig), e
}

func SyncServerIpListSuccessCallBack(responseBody []byte) (o interface{}, err error) {
	Logger.Debug("get all server info:", string(responseBody))

	tmpServerInfo := make([]*config.ServerInfo, 0)

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

func SetDownNode(host string) {
	if host == "" || appConfig == nil {
		return
	}

	if host == appConfig.GetHost() {
		appConfig.SetNextTryConnTime(next_try_connect_period)
	}

	servers.Range(func(k, v interface{}) bool {
		server := v.(*config.ServerInfo)
		// if some node has down then select next node
		if strings.Index(k.(string), host) > -1 {
			server.IsDown = true
			return false
		}
		return true
	})
}

func GetAppConfig(newAppConfig *config.AppConfig) *config.AppConfig {
	if newAppConfig != nil {
		return newAppConfig
	}
	return appConfig
}

func GetServicesConfigUrl(config *config.AppConfig) string {
	return fmt.Sprintf("%sservices/config?appId=%s&ip=%s",
		config.GetHost(),
		url.QueryEscape(config.AppId),
		utils.GetInternal())
}

func GetPlainAppConfig() *config.AppConfig {
	return appConfig
}

func GetServers() *sync.Map {
	return &servers
}

func GetServersLen() int {
	s := GetServers()
	l := 0
	s.Range(func(k, v interface{}) bool {
		l++
		return true
	})
	return l
}

var executeConfigFileOnce sync.Once
var executeGetConfigFile config.ConfigFile

func GetExecuteGetConfigFile() config.ConfigFile {
	executeConfigFileOnce.Do(func() {
		executeGetConfigFile = &json_config.JSONConfigFile{}
	})
	return executeGetConfigFile
}

func Unmarshal(b []byte) (interface{}, error) {
	appConfig := &config.AppConfig{
		Cluster:        default_cluster,
		NamespaceName:  default_namespace,
		IsBackupConfig: true,
	}
	err := json.Unmarshal(b, appConfig)
	if utils.IsNotNil(err) {
		return nil, err
	}

	return appConfig, nil
}
