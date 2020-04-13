package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zouyx/agollo/v3/component/log"
	"github.com/zouyx/agollo/v3/env"
	jsonConfig "github.com/zouyx/agollo/v3/env/config/json"
	"github.com/zouyx/agollo/v3/env/file"
)

//Suffix 默认文件保存类型
const Suffix = ".json"

var (
	//ConfigFileMap 存取namespace文件地址
	ConfigFileMap = make(map[string]string, 1)
	//JsonFileConfig 处理文件的json格式存取
	JsonFileConfig = &jsonConfig.ConfigFile{}
)

func init() {
	file.SetFileHandler(&JSONFileHandler{})
}

//JSONFileHandler 默认备份文件读写
type JSONFileHandler struct {
}

//WriteConfigFile write config to file
func (fileHandler *JSONFileHandler) WriteConfigFile(config *env.ApolloConfig, configPath string) error {
	return JsonFileConfig.Write(config, fileHandler.GetConfigFile(configPath, config.NamespaceName))
}

//GetConfigFile get real config file
func (fileHandler *JSONFileHandler) GetConfigFile(configDir string, namespace string) string {
	fullPath := ConfigFileMap[namespace]
	if fullPath == "" {
		filePath := fmt.Sprintf("%s%s", namespace, Suffix)
		if configDir != "" {
			ConfigFileMap[namespace] = fmt.Sprintf("%s/%s", configDir, filePath)
		} else {
			ConfigFileMap[namespace] = filePath
		}
	}
	return ConfigFileMap[namespace]
}

//LoadConfigFile load config from file
func (fileHandler *JSONFileHandler) LoadConfigFile(configDir string, namespace string) (*env.ApolloConfig, error) {
	configFilePath := fileHandler.GetConfigFile(configDir, namespace)
	log.Info("load config file from :", configFilePath)
	c, e := JsonFileConfig.Load(configFilePath, func(b []byte) (interface{}, error) {
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
