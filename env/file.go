package env

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/zouyx/agollo/v3/component/log"
	jsonConfig "github.com/zouyx/agollo/v3/env/config/json"
)

const suffix = ".json"

var (
	configFileMap  = make(map[string]string, 1)
	jsonFileConfig = &jsonConfig.ConfigFile{}
)

//WriteWithRaw
func WriteWithRaw(f func(config *ApolloConfig, configPath string) error) func(config *ApolloConfig, configPath string) error {
	return func(config *ApolloConfig, configPath string) error {
		filePath := fmt.Sprintf("%s/%s", configPath, config.NamespaceName)
		file, e := os.Create(filePath)
		if e != nil {
			return e
		}
		defer file.Close()
		_, e = file.WriteString(config.Configurations["content"])
		if e != nil {
			return e
		}

		return f(config, configPath)
	}
}

//WriteConfigFile write config to file
func WriteConfigFile(config *ApolloConfig, configPath string) error {
	return jsonFileConfig.Write(config, GetConfigFile(configPath, config.NamespaceName))
}

//GetConfigFile get real config file
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
	log.Info("load config file from :", configFilePath)
	c, e := jsonFileConfig.Load(configFilePath, func(b []byte) (interface{}, error) {
		config := &ApolloConfig{}
		e := json.NewDecoder(bytes.NewBuffer(b)).Decode(config)
		return config, e
	})

	if c == nil || e != nil {
		log.Errorf("loadConfigFile fail,error:", e)
		return nil, e
	}

	return c.(*ApolloConfig), e
}
