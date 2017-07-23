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

const(
	SUCCESS=200
	NOT_MODIFIED=304
)

type ApolloConfig struct {
	AppId string `json:"appId"`
	Cluster string `json:"cluster"`
	NamespaceName string `json:"namespaceName"`
	Configurations map[string]interface{} `json:"configurations"`
	ReleaseKey string `json:"releaseKey"`
	sync.RWMutex
}

func CreateApolloConfigWithJson(b []byte) (*ApolloConfig,error) {
	apolloConfig:=&ApolloConfig{}
	err:=json.Unmarshal(b,apolloConfig)
	if IsNotNil(err) {
		return nil,err
	}
	return apolloConfig,nil
}