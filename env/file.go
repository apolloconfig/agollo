package env

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	. "github.com/zouyx/agollo/v2/component/log"
)

const suffix = ".json"

var configFileMap = make(map[string]string, 1)

//write config to file
func WriteConfigFile(config *ApolloConfig, configPath string) error {
	if config == nil {
		Logger.Error("apollo config is null can not write backup file")
		return errors.New("apollo config is null can not write backup file")
	}
	file, e := os.Create(GetConfigFile(configPath, config.NamespaceName))
	if e != nil {
		Logger.Errorf("writeConfigFile fail,error:", e)
		return e
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(config)
}

//get real config file
func GetConfigFile(configDir string, namespace string) string {
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
func LoadConfigFile(configDir string, namespace string) (*ApolloConfig, error) {
	configFilePath := GetConfigFile(configDir, namespace)
	Logger.Info("load config file from :", configFilePath)
	file, e := os.Open(configFilePath)
	if e != nil {
		Logger.Errorf("loadConfigFile fail,error:", e)
		return nil, e
	}
	defer file.Close()
	config := &ApolloConfig{}
	e = json.NewDecoder(file).Decode(config)

	if e != nil {
		Logger.Errorf("loadConfigFile fail,error:", e)
		return nil, e
	}

	return config, e
}
