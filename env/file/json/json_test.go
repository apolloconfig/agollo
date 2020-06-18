package json

import (
	"encoding/json"
	"github.com/zouyx/agollo/v3/utils"
	"os"
	"testing"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v3/env"
	"github.com/zouyx/agollo/v3/extension"
)

func TestJSONFileHandler_WriteConfigFile(t *testing.T) {
	extension.SetFileHandler(&jsonFileHandler{})
	configPath := ""
	jsonStr := `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application",
  "configurations": {
    "key1":"value1",
    "key2":"value2",
    "test": [1, 2]
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

	config, err := createApolloConfigWithJSON([]byte(jsonStr))
	os.Remove(extension.GetFileHandler().GetConfigFile(configPath, config.NamespaceName))

	Assert(t, err, NilVal())
	e := extension.GetFileHandler().WriteConfigFile(config, configPath)
	Assert(t, e, NilVal())
}

func TestJSONFileHandler_LoadConfigFile(t *testing.T) {
	extension.SetFileHandler(&jsonFileHandler{})
	jsonStr := `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application",
  "configurations": {
    "key1":"value1",
    "key2":"value2",
    "test": [1, 2]
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

	config, err := createApolloConfigWithJSON([]byte(jsonStr))

	Assert(t, err, NilVal())
	newConfig, e := extension.GetFileHandler().LoadConfigFile("", config.NamespaceName)

	t.Log(newConfig)
	Assert(t, e, NilVal())
	Assert(t, config.AppID, Equal(newConfig.AppID))
	Assert(t, config.ReleaseKey, Equal(newConfig.ReleaseKey))
	Assert(t, config.Cluster, Equal(newConfig.Cluster))
	Assert(t, config.NamespaceName, Equal(newConfig.NamespaceName))
}

func createApolloConfigWithJSON(b []byte) (*env.ApolloConfig, error) {
	apolloConfig := &env.ApolloConfig{}
	err := json.Unmarshal(b, apolloConfig)
	if utils.IsNotNil(err) {
		return nil, err
	}
	return apolloConfig, nil
}
