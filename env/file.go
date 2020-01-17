package env

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/zouyx/agollo/v2/component/log"
	"github.com/zouyx/agollo/v2/env/config/json_config"
)

const suffix = ".json"

var (
	configFileMap  = make(map[string]string, 1)
	jsonFileConfig = &json_config.JSONConfigFile{}
)

//write config to file
func WriteConfigFile(config *ApolloConfig, configPath string) error {
	return jsonFileConfig.Write(config, GetConfigFile(configPath, config.NamespaceName))
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

//LoadConfigFile load config from file
func LoadConfigFile(configDir string, namespace string) (*ApolloConfig, error) {
	configFilePath := GetConfigFile(configDir, namespace)
	Logger.Info("load config file from :", configFilePath)
	c, e := jsonFileConfig.Load(configFilePath, func(b []byte) (interface{}, error) {
		config := &ApolloConfig{}
		e := json.NewDecoder(bytes.NewBuffer(b)).Decode(config)
		return config, e
	})

	if c == nil || e != nil {
		Logger.Errorf("loadConfigFile fail,error:", e)
		return nil, e
	}

	return c.(*ApolloConfig), e
}
