package agollo

import (
	"encoding/json"
	"strings"
	"sync"
)

const (
	comma = ","
)

type AbsComponent interface {
	Start()
}

func StartRefreshConfig(component AbsComponent) {
	component.Start()
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
	Namespaces map[string]int64
}

func (this *ApolloConfig) init(appConfig *AppConfig)  {
	this.AppId = appConfig.AppId
	this.Cluster = appConfig.Cluster
	this.NamespaceName = appConfig.NamespaceName
	this.Namespaces=make(map[string]int64,1)

	namespaces := strings.Split(appConfig.NamespaceName, comma)
	for _, v := range namespaces {
		this.Namespaces[v]=default_notification_id
	}
}

func createApolloConfigWithJson(b []byte) (*ApolloConfig, error) {
	apolloConfig := &ApolloConfig{}
	err := json.Unmarshal(b, apolloConfig)
	if isNotNil(err) {
		return nil, err
	}
	return apolloConfig, nil
}
