package dto

import (
	"encoding/json"
	"github.com/zouyx/agollo/utils/objectutils"
	"sync"
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
	if objectutils.IsNotNil(err) {
		return nil,err
	}
	return apolloConfig,nil
}