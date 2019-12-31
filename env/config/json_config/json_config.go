package json_config

import (
	"encoding/json"
	"errors"
	"github.com/zouyx/agollo/v2/env/config"
	"io/ioutil"

	"github.com/zouyx/agollo/v2/utils"
)

var (
	default_cluster   = "default"
	default_namespace = "application"
)

type JSONConfigFile struct {

} 

func(t *JSONConfigFile) LoadJsonConfig(fileName string) (*config.AppConfig, error) {
	fs, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, errors.New("Fail to read config file:" + err.Error())
	}

	appConfig, loadErr := t.Unmarshal(string(fs))

	if utils.IsNotNil(loadErr) {
		return nil, errors.New("Load Json Config fail:" + loadErr.Error())
	}

	return appConfig, nil
}

func(t *JSONConfigFile) Unmarshal(str string) (*config.AppConfig, error) {
	appConfig := &config.AppConfig{
		Cluster:        default_cluster,
		NamespaceName:  default_namespace,
		IsBackupConfig: true,
	}
	err := json.Unmarshal([]byte(str), appConfig)
	if utils.IsNotNil(err) {
		return nil, err
	}

	return appConfig, nil
}
