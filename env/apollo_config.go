package env

import (
	"encoding/json"
	"github.com/zouyx/agollo/v2/utils"
	"sync"
)

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
