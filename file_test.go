package agollo

import (
	. "github.com/tevid/gohamcrest"
	"os"
	"testing"
)

func TestWriteConfigFile(t *testing.T) {
	configPath := ""
	os.Remove(getConfigFile(configPath))
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

	config, err := createApolloConfigWithJson([]byte(jsonStr))

	isNil(err)
	e := writeConfigFile(config, configPath)
	isNil(e)
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

	config, err := createApolloConfigWithJson([]byte(jsonStr))

	isNil(err)
	newConfig, e := loadConfigFile("")

	t.Log(newConfig)
	isNil(e)
	Assert(t, config.AppId, Equal(newConfig.AppId))
	Assert(t, config.ReleaseKey, Equal(newConfig.ReleaseKey))
	Assert(t, config.Cluster, Equal(newConfig.Cluster))
	Assert(t, config.NamespaceName, Equal(newConfig.NamespaceName))
}
