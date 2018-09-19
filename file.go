package agollo

import (
	"encoding/json"
	"fmt"
	"os"
)

const FILE = "apolloConfig.json"
var configFile=""

//write config to file
func writeConfigFile(config *ApolloConfig,configPath string)error{
	file, e := os.Create(getConfigFile(configPath))
	defer  file.Close()
	if e!=nil{
		logger.Errorf("writeConfigFile fail,error:",e)
		return e
	}

	return json.NewEncoder(file).Encode(config)
}

//get real config file
func getConfigFile(configDir string) string {
	if configFile == "" {
		if configDir!="" {
			configFile=fmt.Sprintf("%s/%s",configDir,FILE)
		}else{
			configFile=FILE
		}

	}
	return configFile
}

//load config from file
func loadConfigFile(configDir string) (*ApolloConfig,error){
	configFilePath := getConfigFile(configDir)
	logger.Info("load config file from :",configFilePath)
	file, e := os.Open(configFilePath)
	defer file.Close()
	if e!=nil{
		logger.Errorf("loadConfigFile fail,error:",e)
		return nil,e
	}
	config:=&ApolloConfig{}
	e=json.NewDecoder(file).Decode(config)

	if e!=nil{
		logger.Errorf("loadConfigFile fail,error:",e)
		return nil,e
	}

	return config,e
}