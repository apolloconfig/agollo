package agollo

import (
	"sync"
	"encoding/json"
)

type AbsComponent interface {
	Start()
}


func StartRefreshConfig(component AbsComponent)  {
	component.Start()
}

type ApolloConnConfig struct {
	AppId string `json:"appId"`
	Cluster string `json:"cluster"`
	NamespaceName string `json:"namespaceName"`
	ReleaseKey string `json:"releaseKey"`
	sync.RWMutex
}

type ApolloConfig struct {
	ApolloConnConfig
	Configurations map[string]string `json:"configurations"`
}

func createApolloConfigWithJson(b []byte) (*ApolloConfig,error) {
	apolloConfig:=&ApolloConfig{}
	err:=json.Unmarshal(b,apolloConfig)
	if isNotNil(err) {
		return nil,err
	}
	return apolloConfig,nil
}