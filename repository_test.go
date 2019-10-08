package agollo

import (
	"encoding/json"
	. "github.com/tevid/gohamcrest"
	"testing"
	"time"
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

	updateApolloConfigCache(configs, expireTime)

	return configs
}

func TestTouchApolloConfigCache(t *testing.T) {
	createMockApolloConfig(10)

	time.Sleep(5 * time.Second)
	checkCacheLeft(t, 5)

	updateApolloConfigCacheTime(10)

	checkCacheLeft(t, 10)
}

func checkCacheLeft(t *testing.T, excepted uint32) {
	it := apolloConfigCache.NewIterator()
	for i := int64(0); i < apolloConfigCache.EntryCount(); i++ {
		entry := it.Next()
		left, _ := apolloConfigCache.TTL(entry.Key)
		Assert(t, true, Equal(left == uint32(excepted)))
	}
}

func TestUpdateApolloConfigNull(t *testing.T) {
	time.Sleep(1 * time.Second)
	var currentConfig *ApolloConnConfig
	currentJson, err := json.Marshal(currentConnApolloConfig)
	Assert(t, err,NilVal())

	t.Log("currentJson:", string(currentJson))

	Assert(t, false, Equal(string(currentJson) == ""))

	json.Unmarshal(currentJson, &currentConfig)

	Assert(t, currentConfig,NotNilVal())

	updateApolloConfig(nil, true)

	currentConnApolloConfig.l.RLock()
	defer currentConnApolloConfig.l.RUnlock()
	config := currentConnApolloConfig.config

	//make sure currentConnApolloConfig was not modified
	//Assert(t, currentConfig.NamespaceName, config.NamespaceName)
	//Assert(t, currentConfig.AppId, config.AppId)
	//Assert(t, currentConfig.Cluster, config.Cluster)
	//Assert(t, currentConfig.ReleaseKey, config.ReleaseKey)
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "test", Equal(config.AppId))
	Assert(t, "dev", Equal(config.Cluster))
	Assert(t, "", Equal(config.ReleaseKey))

}

func TestGetApolloConfigCache(t *testing.T) {
	cache := GetApolloConfigCache()
	Assert(t, cache,NotNilVal())
}

func TestGetConfigValueTimeout(t *testing.T) {
	expireTime := 5
	configs := createMockApolloConfig(expireTime)

	for k, v := range configs {
		Assert(t, v, Equal(getValue(k)))
	}

	time.Sleep(time.Duration(expireTime) * time.Second)

	for k := range configs {
		Assert(t, "", Equal(getValue(k)))
	}
}

func TestGetConfigValueNullApolloConfig(t *testing.T) {
	//clear Configurations
	apolloConfigCache.Clear()

	//test getValue
	value := getValue("joe")

	Assert(t, empty, Equal(value))

	//test GetStringValue
	defaultValue := "j"

	//test default
	v := GetStringValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

	createMockApolloConfig(configCacheExpireTime)
}

func TestGetIntValue(t *testing.T) {
	defaultValue := 100000

	//test default
	v := GetIntValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

	//normal value
	v = GetIntValue("int", defaultValue)

	Assert(t, 1, Equal(v))

	//error type
	v = GetIntValue("float", defaultValue)

	Assert(t, defaultValue, Equal(v))
}

func TestGetFloatValue(t *testing.T) {
	defaultValue := 100000.1

	//test default
	v := GetFloatValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

	//normal value
	v = GetFloatValue("float", defaultValue)

	Assert(t, 190.3, Equal(v))

	//error type
	v = GetFloatValue("int", defaultValue)

	Assert(t, float64(1), Equal(v))
}

func TestGetBoolValue(t *testing.T) {
	defaultValue := false

	//test default
	v := GetBoolValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

	//normal value
	v = GetBoolValue("bool", defaultValue)

	Assert(t, true, Equal(v))

	//error type
	v = GetBoolValue("float", defaultValue)

	Assert(t, defaultValue, Equal(v))
}

func TestGetStringValue(t *testing.T) {
	defaultValue := "j"

	//test default
	v := GetStringValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

	//normal value
	v = GetStringValue("string", defaultValue)

	Assert(t, "value", Equal(v))
}
