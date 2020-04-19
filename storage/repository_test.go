package storage

import (
	"strings"
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"
	_ "github.com/zouyx/agollo/v3/agcache/memory"
	"github.com/zouyx/agollo/v3/env"
	_ "github.com/zouyx/agollo/v3/env/file/json"
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
	apolloConfig.AppID = "test"
	apolloConfig.Cluster = "dev"
	apolloConfig.Configurations = configurations
	UpdateApolloConfig(apolloConfig, true)

	currentConnApolloConfig := env.GetCurrentApolloConfig()
	config := currentConnApolloConfig[defaultNamespace]

	Assert(t, config, NotNilVal())
	Assert(t, defaultNamespace, Equal(config.NamespaceName))
	Assert(t, apolloConfig.AppID, Equal(config.AppID))
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

	s = config.GetStringValue("s", "s")
	Assert(t, s, Equal("s"))

	//int
	i := config.GetIntValue("int", 2)
	Assert(t, i, Equal(int(1)))
	i = config.GetIntValue("i", 2)
	Assert(t, i, Equal(int(2)))

	//float
	f := config.GetFloatValue("float", 2)
	Assert(t, f, Equal(float64(1)))
	f = config.GetFloatValue("f", 2)
	Assert(t, f, Equal(float64(2)))

	//bool
	b := config.GetBoolValue("bool", false)
	Assert(t, b, Equal(true))

	b = config.GetBoolValue("b", false)
	Assert(t, b, Equal(false))

	//content
	content := config.GetContent(Properties)
	hasFloat := strings.Contains(content, "float=1")
	Assert(t, hasFloat, Equal(true))

	hasInt := strings.Contains(content, "int=1")
	Assert(t, hasInt, Equal(true))

	hasString := strings.Contains(content, "string=string")
	Assert(t, hasString, Equal(true))

	hasBool := strings.Contains(content, "bool=true")
	Assert(t, hasBool, Equal(true))
}
