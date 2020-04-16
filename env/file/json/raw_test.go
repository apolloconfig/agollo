package json

import (
	"os"
	"testing"

	"github.com/zouyx/agollo/v3/env"
	"github.com/zouyx/agollo/v3/extension"

	. "github.com/tevid/gohamcrest"
)

func TestRawHandler_WriteConfigFile(t *testing.T) {
	extension.SetFileHandler(&RawHandler{})
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
	os.Remove(extension.GetFileHandler().GetConfigFile(configPath, config.NamespaceName))

	Assert(t, err, NilVal())
	e := extension.GetFileHandler().WriteConfigFile(config, configPath)
	Assert(t, e, NilVal())
}
