package defaultfile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zouyx/agollo/v3/component/log"
	"github.com/zouyx/agollo/v3/env"
	"github.com/zouyx/agollo/v3/env/filehandler"
)

func init() {
	filehandler.SetFileHandler(&DefaultFile{})
}

//DefaultFile 默认备份文件读写
type DefaultFile struct {
}

//WriteConfigFile write config to file
func (fileHandler *DefaultFile) WriteConfigFile(config *env.ApolloConfig, configPath string) error {
	return filehandler.JsonFileConfig.Write(config, fileHandler.GetConfigFile(configPath, config.NamespaceName))
}

//GetConfigFile get real config file
func (fileHandler *DefaultFile) GetConfigFile(configDir string, namespace string) string {
	fullPath := filehandler.ConfigFileMap[namespace]
	if fullPath == "" {
		filePath := fmt.Sprintf("%s%s", namespace, filehandler.Suffix)
		if configDir != "" {
			filehandler.ConfigFileMap[namespace] = fmt.Sprintf("%s/%s", configDir, filePath)
		} else {
			filehandler.ConfigFileMap[namespace] = filePath
		}
	}
	return filehandler.ConfigFileMap[namespace]
}

//LoadConfigFile load config from file
func (fileHandler *DefaultFile) LoadConfigFile(configDir string, namespace string) (*env.ApolloConfig, error) {
	configFilePath := fileHandler.GetConfigFile(configDir, namespace)
	log.Info("load config file from :", configFilePath)
	c, e := filehandler.JsonFileConfig.Load(configFilePath, func(b []byte) (interface{}, error) {
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
