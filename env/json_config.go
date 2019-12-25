package env

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/zouyx/agollo/v2/utils"
)

var (
	default_cluster   = "default"
	default_namespace = "application"
)

func loadJsonConfig(fileName string) (*AppConfig, error) {
	fs, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, errors.New("Fail to read config file:" + err.Error())
	}

	appConfig, loadErr := CreateAppConfigWithJson(string(fs))

	if utils.IsNotNil(loadErr) {
		return nil, errors.New("Load Json Config fail:" + loadErr.Error())
	}

	return appConfig, nil
}

func CreateAppConfigWithJson(str string) (*AppConfig, error) {
	appConfig := &AppConfig{
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
