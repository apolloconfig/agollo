package main

import (
	"fmt"

	"github.com/qshuai/agollo/v4"
	"github.com/qshuai/agollo/v4/env/config"
)

func main() {
	config, err := agollo.StartWithConfig(func() (*config.AppConfig, error) {
		return &config.AppConfig{
			AppID:             "jidugokitdemo-key",
			Cluster:           "default",
			NamespaceName:     "default.yaml",
			IsBackupConfig:    false,
			IP:                "http://10.80.0.17:8080",
			BackupConfigPath:  "",
			Secret:            "",
			Label:             "",
			SyncServerTimeout: 0,
			MustStart:         true,
		}, nil
	})
	if err != nil {
		panic(err)
	}

	config.GetConfigCache("default.yaml").Range(func(key, value interface{}) bool {
		fmt.Println(key, value)
		return true
	})

	err = config.AddNamespace("golang.yaml")
	if err != nil {
		panic(err)
	}

	config.GetConfigCache("golang.yaml").Range(func(key, value interface{}) bool {
		fmt.Println(key, value)
		return true
	})
}
