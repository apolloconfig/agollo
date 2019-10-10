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
}

func splitNamespaces(namespacesStr string,callback func(namespace string))map[string]int64{
	namespaces:=make(map[string]int64,1)
	split := strings.Split(namespacesStr, comma)
	for _, namespace := range split {
		callback(namespace)
		namespaces[namespace]=default_notification_id
	}
	return namespaces
}

func (this *ApolloConfig) init(appConfig *AppConfig,namespace string) {
	this.AppId = appConfig.AppId
	this.Cluster = appConfig.Cluster
	this.NamespaceName = namespace
}

func createApolloConfigWithJson(b []byte) (*ApolloConfig, error) {
	apolloConfig := &ApolloConfig{}
	err := json.Unmarshal(b, apolloConfig)
	if isNotNil(err) {
		return nil, err
	}
	return apolloConfig, nil
}
