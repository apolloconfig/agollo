package json

import (
	"os"
	"testing"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v3/extension"
)

func TestRawHandler_WriteConfigFile(t *testing.T) {
	extension.SetFileHandler(&rawFileHandler{})
	configPath := ""
	jsonStr := `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application.json",
  "configurations": {
    "key1":"value1",
    "key2":"value2",
    "test": ["a", "b"]
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

	config, err := createApolloConfigWithJSON([]byte(jsonStr))
	os.Remove(extension.GetFileHandler().GetConfigFile(configPath, config.NamespaceName))

	Assert(t, err, NilVal())
	e := extension.GetFileHandler().WriteConfigFile(config, configPath)
	Assert(t, e, NilVal())
}

func TestRawHandler_WriteConfigFileWithContent(t *testing.T) {
	extension.SetFileHandler(&rawFileHandler{})
	configPath := ""
	jsonStr := `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application.json",
  "configurations": {
    "content":"a: value1"
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

	config, err := createApolloConfigWithJSON([]byte(jsonStr))
	Assert(t, err, NilVal())
	os.Remove(extension.GetFileHandler().GetConfigFile(configPath, config.NamespaceName))

	Assert(t, err, NilVal())
	e := extension.GetFileHandler().WriteConfigFile(config, configPath)
	Assert(t, e, NilVal())
}

func TestGetRawFileHandler(t *testing.T) {
	handler := GetRawFileHandler()
	Assert(t, handler, NotNilVal())

	fileHandler := GetRawFileHandler()
	Assert(t, handler, Equal(fileHandler))
}
