package json

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/zouyx/agollo/v3/extension"

	"github.com/zouyx/agollo/v3/component/log"
	"github.com/zouyx/agollo/v3/env"
	jsonConfig "github.com/zouyx/agollo/v3/env/config/json"
)

//Suffix 默认文件保存类型
const Suffix = ".json"

func init() {
	extension.SetFileHandler(&JSONFileHandler{})
}

var (
	//jsonFileConfig 处理文件的json格式存取
	jsonFileConfig = &jsonConfig.ConfigFile{}
	//configFileMap 存取namespace文件地址
	configFileMap = make(map[string]string, 1)
)

//JSONFileHandler 默认备份文件读写
type JSONFileHandler struct {
}

//WriteConfigFile write config to file
func (fileHandler *JSONFileHandler) WriteConfigFile(config *env.ApolloConfig, configPath string) error {
	return jsonFileConfig.Write(config, fileHandler.GetConfigFile(configPath, config.NamespaceName))
}

//GetConfigFile get real config file
func (fileHandler *JSONFileHandler) GetConfigFile(configDir string, namespace string) string {
	fullPath := configFileMap[namespace]
	if fullPath == "" {
		filePath := fmt.Sprintf("%s%s", namespace, Suffix)
		if configDir != "" {
			configFileMap[namespace] = fmt.Sprintf("%s/%s", configDir, filePath)
		} else {
			configFileMap[namespace] = filePath
		}
	}
	return configFileMap[namespace]
}

//LoadConfigFile load config from file
func (fileHandler *JSONFileHandler) LoadConfigFile(configDir string, namespace string) (*env.ApolloConfig, error) {
	configFilePath := fileHandler.GetConfigFile(configDir, namespace)
	log.Info("load config file from :", configFilePath)
	c, e := jsonFileConfig.Load(configFilePath, func(b []byte) (interface{}, error) {
		config := &env.ApolloConfig{}
		e := json.NewDecoder(bytes.NewBuffer(b)).Decode(config)
		return config, e
	})

	if c == nil || e != nil {
		log.Errorf("loadConfigFile fail,error:", e)
		return nil, e
	}

	return c.(*env.ApolloConfig), e
}
