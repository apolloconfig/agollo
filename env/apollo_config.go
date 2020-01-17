package env

import (
	"encoding/json"
	"sync"

	"github.com/zouyx/agollo/v2/utils"
)

var (
	currentConnApolloConfig = &currentApolloConfig{
		configs: make(map[string]*ApolloConnConfig, 1),
	}
)

type currentApolloConfig struct {
	l       sync.RWMutex
	configs map[string]*ApolloConnConfig
}

//SetCurrentApolloConfig 设置apollo配置
func SetCurrentApolloConfig(namespace string, connConfig *ApolloConnConfig) {
	currentConnApolloConfig.l.Lock()
	defer currentConnApolloConfig.l.Unlock()

	currentConnApolloConfig.configs[namespace] = connConfig
}

//GetCurrentApolloConfig 获取Apollo链接配置
func GetCurrentApolloConfig() map[string]*ApolloConnConfig {
	currentConnApolloConfig.l.RLock()
	defer currentConnApolloConfig.l.RUnlock()

	return currentConnApolloConfig.configs
}

//GetCurrentApolloConfigReleaseKey 获取release key
func GetCurrentApolloConfigReleaseKey(namespace string) string {
	currentConnApolloConfig.l.RLock()
	defer currentConnApolloConfig.l.RUnlock()
	config := currentConnApolloConfig.configs[namespace]
	if config == nil {
		return utils.Empty
	}

	return config.ReleaseKey
}

//ApolloConnConfig apollo链接配置
type ApolloConnConfig struct {
	AppID         string `json:"appId"`
	Cluster       string `json:"cluster"`
	NamespaceName string `json:"namespaceName"`
	ReleaseKey    string `json:"releaseKey"`
	sync.RWMutex
}

//ApolloConfig apollo配置
type ApolloConfig struct {
	ApolloConnConfig
	Configurations map[string]string `json:"configurations"`
}

//Init 初始化
func (a *ApolloConfig) Init(appId string, cluster string, namespace string) {
	a.AppID = appId
	a.Cluster = cluster
	a.NamespaceName = namespace
}

//CreateApolloConfigWithJson 使用json配置转换成apolloconfig
func CreateApolloConfigWithJson(b []byte) (*ApolloConfig, error) {
	apolloConfig := &ApolloConfig{}
	err := json.Unmarshal(b, apolloConfig)
	if utils.IsNotNil(err) {
		return nil, err
	}
	return apolloConfig, nil
}
