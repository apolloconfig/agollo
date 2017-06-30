package repository

import "sync"

var configRepository *ConfigRepository

func init() {
	configRepository=&ConfigRepository{
		//configMap:make(map[string]interface{}),
	}
}

type ConfigRepository struct {
	configMap map[string]interface{}
	sync.RWMutex
}

func UpdateConfig(configMap map[string]interface{})  {
	configRepository.Lock()
	defer configRepository.Unlock()

	configRepository.configMap=configMap
}

func GetConfig() *ConfigRepository{
	return configRepository
}