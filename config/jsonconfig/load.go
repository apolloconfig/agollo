package jsonconfig

import (
	"github.com/zouyx/agollo/dto"
	"io/ioutil"
	"github.com/zouyx/agollo/utils/objectutils"
	"github.com/cihub/seelog"
)

func Load() *dto.ApolloConfig {
	fs, err := ioutil.ReadFile("app.properties")
	if err != nil {
		panic("faile to read config dir:" + err.Error())
	}

	apolloConfig,loadErr:=dto.CreateApolloConfigWithJson(string(fs))

	if objectutils.IsNotNil(loadErr){
		seelog.Errorf("Load Json Config is fail:%s",loadErr)
	}

	return apolloConfig
}