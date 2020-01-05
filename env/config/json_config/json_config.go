package json_config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	. "github.com/zouyx/agollo/v2/component/log"
	"github.com/zouyx/agollo/v2/utils"
)

type JSONConfigFile struct {
}

func (t *JSONConfigFile) Load(fileName string, unmarshal func([]byte) (interface{}, error)) (interface{}, error) {
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

func (t *JSONConfigFile) Write(content interface{}, configPath string) error {
	if content == nil {
		Logger.Error("content is null can not write backup file")
		return errors.New("content is null can not write backup file")
	}
	file, e := os.Create(configPath)
	if e != nil {
		Logger.Errorf("writeConfigFile fail,error:", e)
		return e
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(content)
}
