package storage

import (
	"testing"
	"time"

	"github.com/zouyx/agollo/v2/env"

	. "github.com/tevid/gohamcrest"
)

//init param
func init() {
}

func TestUpdateApolloConfigNull(t *testing.T) {
	time.Sleep(1 * time.Second)

	apolloConfig := &env.ApolloConfig{}
	apolloConfig.NamespaceName = defaultNamespace
	apolloConfig.AppId = "test"
	apolloConfig.Cluster = "dev"
	UpdateApolloConfig(apolloConfig, true)

	currentConnApolloConfig := env.GetCurrentApolloConfig()
	config := currentConnApolloConfig[defaultNamespace]

	Assert(t, config, NotNilVal())
	Assert(t, defaultNamespace, Equal(config.NamespaceName))
	Assert(t, apolloConfig.AppId, Equal(config.AppId))
	Assert(t, apolloConfig.Cluster, Equal(config.Cluster))
	Assert(t, "", Equal(config.ReleaseKey))

}

func TestGetApolloConfigCache(t *testing.T) {
	cache := GetApolloConfigCache()
	Assert(t, cache, NotNilVal())
}
