package repository

import "sync"

var configRepository *ConfigRepository

func init() {
	configRepository=&ConfigRepository{}
}

type ConfigRepository struct {
	configMap map[string]interface{}
	*sync.RWMutex
}

func UpdateConfig(configMap map[string]interface{})  {
	configRepository.Lock()
	defer configRepository.Unlock()

	configRepository.configMap=configMap
}