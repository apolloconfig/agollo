package env

import (
	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v2/utils"
	"testing"
)

func TestSetCurrentApolloConfig(t *testing.T) {
	Assert(t, currentConnApolloConfig.configs, NotNilVal())
	currentConnApolloConfig.configs["a"] = &ApolloConnConfig{
		AppId:      "a",
		ReleaseKey: "releaseKey",
	}
}

func TestGetCurrentApolloConfig(t *testing.T) {
	Assert(t, currentConnApolloConfig.configs, NotNilVal())
	config := GetCurrentApolloConfig()["a"]
	Assert(t, config, NotNilVal())
	Assert(t, config.AppId, Equal("a"))
}

func TestGetCurrentApolloConfigReleaseKey(t *testing.T) {
	Assert(t, currentConnApolloConfig.configs, NotNilVal())
	key := GetCurrentApolloConfigReleaseKey("b")
	Assert(t, key, Equal(utils.Empty))

	key = GetCurrentApolloConfigReleaseKey("a")
	Assert(t, key, Equal("releaseKey"))
}
