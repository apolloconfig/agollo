package agollo

import (
	"io/ioutil"
	"github.com/cihub/seelog"
)

func LoadJsonConfig() *AppConfig {
	fs, err := ioutil.ReadFile("app.properties")
	if err != nil {
		panic("faile to read config dir:" + err.Error())
	}

	appConfig,loadErr:=CreateAppConfigWithJson(string(fs))

	if IsNotNil(loadErr){
		seelog.Errorf("Load Json Config is fail:%s",loadErr)
	}

	return appConfig
}
