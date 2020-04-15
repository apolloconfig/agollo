package json

import (
	"fmt"
	"github.com/zouyx/agollo/v3/env"
	"os"
)

//RawHandler 写入备份文件时，同时写入原始内容和namespace类型
type RawHandler struct {
	*JSONFileHandler
}

func writeWithRaw(config *env.ApolloConfig, configDir string) error {
	filePath := ""
	if configDir != "" {
		filePath = fmt.Sprintf("%s/%s", configDir, config.NamespaceName)
	} else {
		filePath = config.NamespaceName
	}

	file, e := os.Create(filePath)
	if e != nil {
		return e
	}
	defer file.Close()
	_, e = file.WriteString(config.Configurations["content"])
	if e != nil {
		return e
	}
	return nil
}

//WriteConfigFile write config to file
func (fileHandler *RawHandler) WriteConfigFile(config *env.ApolloConfig, configPath string) error {
	writeWithRaw(config, configPath)
	return jsonFileConfig.Write(config, fileHandler.GetConfigFile(configPath, config.NamespaceName))
}
