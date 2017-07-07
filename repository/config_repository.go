package repository

import (
	"github.com/zouyx/agollo/dto"
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