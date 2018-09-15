package agollo

import (
	"encoding/json"
	"os"
)

const FILE = "a.json"

func writeConfigFile(config *ApolloConfig)error{
	file, e := os.Create(FILE)
	defer  file.Close()
	if e!=nil{
		return e
	}

	return json.NewEncoder(file).Encode(config)
}

func loadConfigFile() (*ApolloConfig,error){
	file, e := os.Open(FILE)
	defer  file.Close()
	if e!=nil{
		return nil,e
	}
	config:=&ApolloConfig{}
	e=json.NewDecoder(file).Decode(config)
	return config,e
}