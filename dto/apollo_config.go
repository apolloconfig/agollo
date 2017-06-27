package dto

import (
	"encoding/json"
	"github.com/zouyx/agollo/utils/objectutils"
)

type ApolloConfig struct {
	AppId string `json:"appId"`
	Cluster string `json:"cluster"`
	NamespaceName string `json:"namespaceName"`
	Configurations map[string]string `json:"-"`
	ReleaseKey string `json:"releaseKey"`
}

func CreateApolloConfig(appId string,
	cluster string,
	namespaceName string,
	releaseKey string) *ApolloConfig {
	apolloConfig:=&ApolloConfig{
		AppId:appId,
		Cluster:cluster,
		NamespaceName:namespaceName,
		ReleaseKey:releaseKey,
	}
	return apolloConfig
}

func CreateApolloConfigWithJson(str string) (*ApolloConfig,error) {
	apolloConfig:=&ApolloConfig{}
	err:=json.Unmarshal([]byte(str),apolloConfig)
	if objectutils.IsNotNil(err) {
		return nil,err
	}
	return apolloConfig,nil
}