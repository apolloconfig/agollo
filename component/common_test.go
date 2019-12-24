package component

import (
	. "github.com/tevid/gohamcrest"
	"testing"
)

func TestCreateApolloConfigWithJson(t *testing.T) {
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

	Assert(t, err, NilVal())
	Assert(t, config, NotNilVal())

	Assert(t, "100004458", Equal(config.AppId))
	Assert(t, "default", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "20170430092936-dee2d58e74515ff3", Equal(config.ReleaseKey))
	Assert(t, "value1", Equal(config.Configurations["key1"]))
	Assert(t, "value2", Equal(config.Configurations["key2"]))

}

func TestCreateApolloConfigWithJsonError(t *testing.T) {
	jsonStr := `jklasdjflasjdfa`

	config, err := createApolloConfigWithJson([]byte(jsonStr))

	Assert(t, err, NotNilVal())
	Assert(t, config, NilVal())
}
