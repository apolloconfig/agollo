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
func getConfigFile(configPath string) string {
	if configFile == "" {
		if configPath!="" {
			configFile=fmt.Sprintf("%s/%s",configPath,FILE)
		}else{
			configFile=FILE
		}

	}
	return configFile
}

//load config from file
func loadConfigFile(configPath string) (*ApolloConfig,error){
	file, e := os.Open(getConfigFile(configPath))
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