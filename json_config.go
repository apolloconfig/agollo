package agollo

import (
	"io/ioutil"
	"encoding/json"
	"errors"
)

var(
	default_cluster="default"
	default_namespace="application"
)

func loadJsonConfig(fileName string) (*AppConfig,error) {
	fs, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil,errors.New("Fail to read config file:" + err.Error())
	}

	appConfig,loadErr:=createAppConfigWithJson(string(fs))

	if isNotNil(loadErr){
		return nil,errors.New("Load Json Config fail:" + loadErr.Error())
	}

	return appConfig,nil
}

func createAppConfigWithJson(str string) (*AppConfig,error) {
	appConfig:=&AppConfig{
		Cluster:default_cluster,
		NamespaceName:default_namespace,
	}
	err:=json.Unmarshal([]byte(str),appConfig)
	if isNotNil(err) {
		return nil,err
	}

	return appConfig,nil
}

