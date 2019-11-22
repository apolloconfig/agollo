package agollo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

const suffix = ".json"

var configFileMap=make(map[string]string,1)

//write config to file
func writeConfigFile(config *ApolloConfig, configPath string) error {
	if config == nil {
		logger.Error("apollo config is null can not write backup file")
		return errors.New("apollo config is null can not write backup file")
	}
	file, e := os.Create(getConfigFile(configPath,config.NamespaceName))
	if e != nil {
		logger.Errorf("writeConfigFile fail,error:", e)
		return e
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(config)
}

//get real config file
func getConfigFile(configDir string,namespace string) string {
	fullPath := configFileMap[namespace]
	if fullPath == "" {
		filePath := fmt.Sprintf("%s%s", namespace, suffix)
		if configDir != "" {
			configFileMap[namespace] = fmt.Sprintf("%s/%s", configDir, filePath)
		} else {
			configFileMap[namespace] = filePath
		}
	}
	return configFileMap[namespace]
}

//load config from file
func loadConfigFile(configDir string,namespace string) (*ApolloConfig, error) {
	configFilePath := getConfigFile(configDir,namespace)
	logger.Info("load config file from :", configFilePath)
	file, e := os.Open(configFilePath)
	if e != nil {
		logger.Errorf("loadConfigFile fail,error:", e)
		return nil, e
	}
	defer file.Close()
	config := &ApolloConfig{}
	e = json.NewDecoder(file).Decode(config)

	if e != nil {
		logger.Errorf("loadConfigFile fail,error:", e)
		return nil, e
	}

	return config, e
}
