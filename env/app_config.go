package env

import (
	"encoding/json"
	"fmt"
	. "github.com/zouyx/agollo/v2/component/log"
	"github.com/zouyx/agollo/v2/env/config"
	jsonConfig "github.com/zouyx/agollo/v2/env/config/json"
	"github.com/zouyx/agollo/v2/utils"
	"net/url"
	"os"
	"strings"
	"sync"
)

const (
	appConfigFile     = "app.properties"
	appConfigFilePath = "AGOLLO_CONF"

	defaultNotificationID = -1
	comma                 = ","

	defaultCluster   = "default"
	defaultNamespace = "application"
)

var (
	//appconfig
	appConfig *config.AppConfig
	//real servers ip
	servers sync.Map

	//next try connect period - 60 second
	nextTryConnectPeriod int64 = 60
)

func init() {
	//init config
	InitFileConfig()
}

//InitFileConfig 使用文件初始化配置
func InitFileConfig() {
	// default use application.properties
	InitConfig(nil)
}

//InitConfig 使用指定配置初始化配置
func InitConfig(loadAppConfig func() (*config.AppConfig, error)) {
	var err error
	//init config file
	appConfig, err = getLoadAppConfig(loadAppConfig)

	if err != nil {
		return
	}
}

//SplitNamespaces 根据namespace字符串分割后，并执行callback函数
func SplitNamespaces(namespacesStr string, callback func(namespace string)) map[string]int64 {
	namespaces := make(map[string]int64, 1)
	split := strings.Split(namespacesStr, comma)
	for _, namespace := range split {
		if callback != nil {
			callback(namespace)
		}
		namespaces[namespace] = defaultNotificationID
	}
	return namespaces
}

// set load app config's function
func getLoadAppConfig(loadAppConfig func() (*config.AppConfig, error)) (*config.AppConfig, error) {
	if loadAppConfig != nil {
		return loadAppConfig()
	}
	configPath := os.Getenv(appConfigFilePath)
	if configPath == "" {
		configPath = appConfigFile
	}
	c, e := GetConfigFileExecutor().Load(configPath, Unmarshal)
	if c == nil {
		return nil, e
	}

	return c.(*config.AppConfig), e
}

//SyncServerIPListSuccessCallBack 同步服务器列表成功后的回调
func SyncServerIPListSuccessCallBack(responseBody []byte) (o interface{}, err error) {
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
		servers.Store(server.HomepageURL, server)
	}
	return
}

//SetDownNode 设置失效节点
func SetDownNode(host string) {
	if host == "" || appConfig == nil {
		return
	}

	if host == appConfig.GetHost() {
		appConfig.SetNextTryConnTime(nextTryConnectPeriod)
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

//GetAppConfig 获取app配置
func GetAppConfig(newAppConfig *config.AppConfig) *config.AppConfig {
	if newAppConfig != nil {
		return newAppConfig
	}
	return appConfig
}

//GetServicesConfigURL 获取服务器列表url
func GetServicesConfigURL(config *config.AppConfig) string {
	return fmt.Sprintf("%sservices/config?appId=%s&ip=%s",
		config.GetHost(),
		url.QueryEscape(config.AppID),
		utils.GetInternal())
}

//GetPlainAppConfig 获取原始配置
func GetPlainAppConfig() *config.AppConfig {
	return appConfig
}

//GetServers 获取服务器数组
func GetServers() *sync.Map {
	return &servers
}

//GetServersLen 获取服务器数组长度
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
var configFileExecutor config.ConfigFile

//GetConfigFileExecutor 获取文件执行器
func GetConfigFileExecutor() config.ConfigFile {
	executeConfigFileOnce.Do(func() {
		configFileExecutor = &jsonConfig.ConfigFile{}
	})
	return configFileExecutor
}

//Unmarshal 反序列化
func Unmarshal(b []byte) (interface{}, error) {
	appConfig := &config.AppConfig{
		Cluster:        defaultCluster,
		NamespaceName:  defaultNamespace,
		IsBackupConfig: true,
	}
	err := json.Unmarshal(b, appConfig)
	if utils.IsNotNil(err) {
		return nil, err
	}

	return appConfig, nil
}
