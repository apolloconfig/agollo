package env

import (
	"os"
	"testing"

	. "github.com/tevid/gohamcrest"
)

func TestWriteConfigFile(t *testing.T) {
	configPath := ""
	jsonStr := `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application",
  "configurations": {
    "key1":"value1",
    "key2":"value2"
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

	config, err := CreateApolloConfigWithJson([]byte(jsonStr))
	os.Remove(GetConfigFile(configPath, config.NamespaceName))

	Assert(t, err, NilVal())
	e := WriteConfigFile(config, configPath)
	Assert(t, e, NilVal())
}

func TestLoadConfigFile(t *testing.T) {
	jsonStr := `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application",
  "configurations": {
    "key1":"value1",
    "key2":"value2"
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

	config, err := CreateApolloConfigWithJson([]byte(jsonStr))

	Assert(t, err, NilVal())
	newConfig, e := LoadConfigFile("", config.NamespaceName)

	t.Log(newConfig)
	Assert(t, e, NilVal())
	Assert(t, config.AppId, Equal(newConfig.AppId))
	Assert(t, config.ReleaseKey, Equal(newConfig.ReleaseKey))
	Assert(t, config.Cluster, Equal(newConfig.Cluster))
	Assert(t, config.NamespaceName, Equal(newConfig.NamespaceName))
}
