package env

import (
	"fmt"
	"os"
	"testing"

	. "github.com/tevid/gohamcrest"
)

func TestWriteWithRaw(t *testing.T) {
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

	config, err := CreateApolloConfigWithJSON([]byte(jsonStr))
	filePath := fmt.Sprintf("%s/%s", configPath, config.NamespaceName)
	os.Remove(filePath)
	Assert(t, err, NilVal())
	e := WriteWithRaw(WriteConfigFile)(config, configPath)
	Assert(t, e, NilVal())
}

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

	config, err := CreateApolloConfigWithJSON([]byte(jsonStr))
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

	config, err := CreateApolloConfigWithJSON([]byte(jsonStr))

	Assert(t, err, NilVal())
	newConfig, e := LoadConfigFile("", config.NamespaceName)

	t.Log(newConfig)
	Assert(t, e, NilVal())
	Assert(t, config.AppID, Equal(newConfig.AppID))
	Assert(t, config.ReleaseKey, Equal(newConfig.ReleaseKey))
	Assert(t, config.Cluster, Equal(newConfig.Cluster))
	Assert(t, config.NamespaceName, Equal(newConfig.NamespaceName))
}
