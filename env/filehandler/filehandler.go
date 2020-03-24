package filehandler

import (
	"github.com/zouyx/agollo/v3/env"
	jsonConfig "github.com/zouyx/agollo/v3/env/config/json"
)

const Suffix = ".json"

var file FileHandler

var (
	ConfigFileMap  = make(map[string]string, 1)
	JsonFileConfig = &jsonConfig.ConfigFile{}
)

//FileHandler 备份文件读写
type FileHandler interface {
	WriteConfigFile(config *env.ApolloConfig, configPath string) error
	GetConfigFile(configDir string, namespace string) string
	LoadConfigFile(configDir string, namespace string) (*env.ApolloConfig, error)
}

//SetFileHandler 设置备份文件处理
func SetFileHandler(inFile FileHandler) {
	file = inFile
}

//GetFileHandler 获取备份文件处理
func GetFileHandler() FileHandler {
	return file
}
