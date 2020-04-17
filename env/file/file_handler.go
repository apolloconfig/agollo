package file

import (
	"github.com/zouyx/agollo/v3/env"
)

//FileHandler 备份文件读写
type FileHandler interface {
	WriteConfigFile(config *env.ApolloConfig, configPath string) error
	GetConfigFile(configDir string, namespace string) string
	LoadConfigFile(configDir string, namespace string) (*env.ApolloConfig, error)
}
