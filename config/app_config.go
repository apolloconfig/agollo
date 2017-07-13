package config

import (
	"github.com/zouyx/agollo/dto"
	"github.com/zouyx/agollo/config/jsonconfig"
	"github.com/zouyx/agollo/repository"
)

var (
	appConfig *dto.AppConfig
)

func init() {
	//init config file
	appConfig = jsonconfig.Load()

	go func(appConfig *dto.AppConfig) {
		apolloConfig:=&dto.ApolloConfig{
			AppId:appConfig.AppId,
			Cluster:appConfig.Cluster,
			NamespaceName:appConfig.NamespaceName,
		}

		repository.UpdateApolloConfig(apolloConfig)
	}(appConfig)


}

func GetAppConfig()*dto.AppConfig  {
	return appConfig
}

