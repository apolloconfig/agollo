package json

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/zouyx/agollo/v3/component/log"
	"github.com/zouyx/agollo/v3/utils"
)

//ConfigFile json文件读写
type ConfigFile struct {
}

//Load json文件读
func (t *ConfigFile) Load(fileName string, unmarshal func([]byte) (interface{}, error)) (interface{}, error) {
	fs, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, errors.New("Fail to read config file:" + err.Error())
	}

	config, loadErr := unmarshal(fs)

	if utils.IsNotNil(loadErr) {
		return nil, errors.New("Load Json Config fail:" + loadErr.Error())
	}

	return config, nil
}

//Write json文件写
func (t *ConfigFile) Write(content interface{}, configPath string) error {
	if content == nil {
		log.Error("content is null can not write backup file")
		return errors.New("content is null can not write backup file")
	}
	file, e := os.Create(configPath)
	if e != nil {
		log.Errorf("writeConfigFile fail,error:", e)
		return e
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(content)
}
