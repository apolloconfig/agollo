package agollo

import (
	"encoding/json"
	"os"
)

const FILE = "apolloConfig.json"

func writeConfigFile(config *ApolloConfig,configPath string)error{
	file, e := os.Create(FILE)
	defer  file.Close()
	if e!=nil{
		logger.Errorf("writeConfigFile fail,error:",e)
		return e
	}

	return json.NewEncoder(file).Encode(config)
}

func loadConfigFile(configPath string) (*ApolloConfig,error){
	file, e := os.Open(FILE)
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