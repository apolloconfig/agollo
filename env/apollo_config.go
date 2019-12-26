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

func GetCurrentApolloConfigReleaseKey(namespace string) string {
	currentConnApolloConfig.l.RLock()
	defer currentConnApolloConfig.l.RUnlock()
	config := currentConnApolloConfig.configs[namespace]
	if config == nil {
		return utils.Empty
	}

	return config.ReleaseKey
}

type ApolloConnConfig struct {
	AppId         string `json:"appId"`
	Cluster       string `json:"cluster"`
	NamespaceName string `json:"namespaceName"`
	ReleaseKey    string `json:"releaseKey"`
	sync.RWMutex
}

type ApolloConfig struct {
	ApolloConnConfig
	Configurations map[string]string `json:"configurations"`
}

func (a *ApolloConfig) Init(appId string, cluster string, namespace string) {
	a.AppId = appId
	a.Cluster = cluster
	a.NamespaceName = namespace
}

func CreateApolloConfigWithJson(b []byte) (*ApolloConfig, error) {
	apolloConfig := &ApolloConfig{}
	err := json.Unmarshal(b, apolloConfig)
	if utils.IsNotNil(err) {
		return nil, err
	}
	return apolloConfig, nil
}
