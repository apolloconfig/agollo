package dto

import (
	"encoding/json"
	"github.com/zouyx/agollo/utils/objectutils"
)

type AppConfig struct {
	AppId string `json:"appId"`
	Cluster string `json:"cluster"`
	NamespaceName string `json:"namespaceName"`
	ReleaseKey string `json:"releaseKey"`
	Ip string `json:"ip"`
}

func CreateAppConfigWithJson(str string) (*AppConfig,error) {
	appConfig:=&AppConfig{}
	err:=json.Unmarshal([]byte(str),appConfig)
	if objectutils.IsNotNil(err) {
		return nil,err
	}
	return appConfig,nil
}