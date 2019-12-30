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

func createMockApolloConfig(expireTime int) map[string]string {
	configs := make(map[string]string, 0)
	//string
	configs["string"] = "value"
	//int
	configs["int"] = "1"
	//float
	configs["float"] = "190.3"
	//bool
	configs["bool"] = "true"

	UpdateApolloConfigCache(configs, expireTime, defaultNamespace)

	return configs
}

func TestUpdateApolloConfigNull(t *testing.T) {
	time.Sleep(1 * time.Second)

	apolloConfig := &env.ApolloConfig{
	}
	apolloConfig.NamespaceName=defaultNamespace
	apolloConfig.AppId="test"
	apolloConfig.Cluster="dev"
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

func TestConfig_GetStringValue(t *testing.T) {
	config := GetConfig(defaultNamespace)

	defaultValue := "j"
	//test default
	v := config.GetStringValue("joe", defaultValue)
	Assert(t, defaultValue, Equal(v))

	//normal value
	v = config.GetStringValue("string", defaultValue)

	Assert(t, "value", Equal(v))
}

func TestConfig_GetBoolValue(t *testing.T) {
	defaultValue := false
	config := GetConfig(defaultNamespace)

	//test default
	v := config.GetBoolValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

	//normal value
	v = config.GetBoolValue("bool", defaultValue)

	Assert(t, true, Equal(v))

	//error type
	v = config.GetBoolValue("float", defaultValue)

	Assert(t, defaultValue, Equal(v))
}

func TestConfig_GetFloatValue(t *testing.T) {
	defaultValue := 100000.1
	config := GetConfig(defaultNamespace)

	//test default
	v := config.GetFloatValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

	//normal value
	v = config.GetFloatValue("float", defaultValue)

	Assert(t, 190.3, Equal(v))

	//error type
	v = config.GetFloatValue("int", defaultValue)

	Assert(t, float64(1), Equal(v))
}

func TestConfig_GetIntValue(t *testing.T) {
	defaultValue := 100000
	config := GetConfig(defaultNamespace)

	//test default
	v := config.GetIntValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

	//normal value
	v = config.GetIntValue("int", defaultValue)

	Assert(t, 1, Equal(v))

	//error type
	v = config.GetIntValue("float", defaultValue)

	Assert(t, defaultValue, Equal(v))
}
