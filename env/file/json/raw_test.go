package json

import (
	"github.com/zouyx/agollo/v3/env"
	"github.com/zouyx/agollo/v3/env/file"
	"os"
	"testing"

	. "github.com/tevid/gohamcrest"
)

func TestRawHandler_WriteConfigFile(t *testing.T) {
	file.SetFileHandler(&RawHandler{})
	configPath := ""
	jsonStr := `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application.json",
  "configurations": {
    "key1":"value1",
    "key2":"value2"
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

	config, err := env.CreateApolloConfigWithJSON([]byte(jsonStr))
	os.Remove(file.GetFileHandler().GetConfigFile(configPath, config.NamespaceName))

	Assert(t, err, NilVal())
	e := file.GetFileHandler().WriteConfigFile(config, configPath)
	Assert(t, e, NilVal())
}
