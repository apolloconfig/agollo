package repository

import (
	"github.com/zouyx/agollo/dto"
)

const (
	empty  =""
)

var (
	currentApolloConfig *dto.ApolloConfig
)

func init(){
	currentApolloConfig=&dto.ApolloConfig{}
}

func UpdateApolloConfig(apolloConfig *dto.ApolloConfig)  {
	currentApolloConfig.Lock()
	defer currentApolloConfig.Unlock()
	currentApolloConfig=apolloConfig
}

func GetCurrentApolloConfig()*dto.ApolloConfig  {
	currentApolloConfig.RLock()
	defer currentApolloConfig.RUnlock()
	return currentApolloConfig
}

func getConfigValue(key string) interface{}  {
	if currentApolloConfig==nil ||currentApolloConfig.Configurations==nil  {
		return empty
	}

	currentApolloConfig.RLock()
	defer currentApolloConfig.RUnlock()

	return currentApolloConfig.Configurations[key]
}