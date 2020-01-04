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

	configurations := make(map[string]string)
	configurations["string"] = "string"
	configurations["int"] = "1"
	configurations["float"] = "1"
	configurations["bool"] = "true"

	apolloConfig := &env.ApolloConfig{}
	apolloConfig.NamespaceName = defaultNamespace
	apolloConfig.AppId = "test"
	apolloConfig.Cluster = "dev"
	apolloConfig.Configurations = configurations
	UpdateApolloConfig(apolloConfig, true)

	currentConnApolloConfig := env.GetCurrentApolloConfig()
	config := currentConnApolloConfig[defaultNamespace]

	Assert(t, config, NotNilVal())
	Assert(t, defaultNamespace, Equal(config.NamespaceName))
	Assert(t, apolloConfig.AppId, Equal(config.AppId))
	Assert(t, apolloConfig.Cluster, Equal(config.Cluster))
	Assert(t, "", Equal(config.ReleaseKey))
	Assert(t, len(apolloConfig.Configurations), Equal(4))

}

func TestGetApolloConfigCache(t *testing.T) {
	cache := GetApolloConfigCache()
	Assert(t, cache, NotNilVal())
}

func TestGetDefaultNamespace(t *testing.T) {
	namespace := GetDefaultNamespace()
	Assert(t, namespace, Equal(defaultNamespace))
}

func TestGetConfig(t *testing.T) {
	config := GetConfig(defaultNamespace)
	Assert(t, config, NotNilVal())

	//string
	s := config.GetStringValue("string", "s")
	Assert(t, s, Equal("string"))

	//int
	i := config.GetIntValue("int", 2)
	Assert(t, i, Equal(int(1)))

	//float
	f := config.GetFloatValue("float", 2)
	Assert(t, f, Equal(float64(1)))

	//bool
	b := config.GetBoolValue("bool", false)
	Assert(t, b, Equal(true))
}
