package agollo

import (
	"io/ioutil"
	"encoding/json"
	"errors"
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
	appConfig:=&AppConfig{}
	err:=json.Unmarshal([]byte(str),appConfig)
	if isNotNil(err) {
		return nil,err
	}

	return appConfig,nil
}

