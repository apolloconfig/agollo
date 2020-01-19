package env

import (
	"testing"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v3/utils"
)

func TestSetCurrentApolloConfig(t *testing.T) {
	Assert(t, currentConnApolloConfig.configs, NotNilVal())
	config := &ApolloConnConfig{
		AppID:      "a",
		ReleaseKey: "releaseKey",
	}
	SetCurrentApolloConfig("a", config)
}

func TestGetCurrentApolloConfig(t *testing.T) {
	Assert(t, currentConnApolloConfig.configs, NotNilVal())
	config := GetCurrentApolloConfig()["a"]
	Assert(t, config, NotNilVal())
	Assert(t, config.AppID, Equal("a"))
}

func TestGetCurrentApolloConfigReleaseKey(t *testing.T) {
	Assert(t, currentConnApolloConfig.configs, NotNilVal())
	key := GetCurrentApolloConfigReleaseKey("b")
	Assert(t, key, Equal(utils.Empty))

	key = GetCurrentApolloConfigReleaseKey("a")
	Assert(t, key, Equal("releaseKey"))
}

func TestApolloConfigInit(t *testing.T) {
	config := &ApolloConfig{}
	config.Init("appId", "cluster", "ns")

	Assert(t, config.AppID, Equal("appId"))
	Assert(t, config.Cluster, Equal("cluster"))
	Assert(t, config.NamespaceName, Equal("ns"))
}
