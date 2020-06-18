package agollo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v3/component/notify"
	"github.com/zouyx/agollo/v3/env"
	_ "github.com/zouyx/agollo/v3/env/file/json"
	"github.com/zouyx/agollo/v3/extension"
	"github.com/zouyx/agollo/v3/storage"
)

const testDefaultNamespace = "application"

//init param
func init() {
}

func createMockApolloConfig(expireTime int) map[string]interface{}{
	configs := make(map[string]interface{}, 0)
	//string
	configs["string"] = "value"
	//int
	configs["int"] = "1"
	//float
	configs["float"] = "190.3"
	//bool
	configs["bool"] = "true"
	//string slice
	configs["stringSlice"] = []string{"1", "2"}

	//int slice
	configs["intSlice"] = []int{1, 2}

	storage.UpdateApolloConfigCache(configs, expireTime, storage.GetDefaultNamespace())

	return configs
}

func TestGetConfigValueNullApolloConfig(t *testing.T) {
	createMockApolloConfig(120)
	//clear Configurations
	defaultConfigCache := GetDefaultConfigCache()
	defaultConfigCache.Clear()

	//test getValue
	value := GetValue("joe")

	Assert(t, "", Equal(value))

	//test GetStringValue
	defaultValue := "j"

	//test default
	v := GetStringValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

}

func TestGetIntValue(t *testing.T) {
	createMockApolloConfig(120)
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

func TestGetIntSliceValue(t *testing.T) {
	createMockApolloConfig(120)
	defaultValue := []int{100}

	//test default
	v := GetIntSliceValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

	//normal value
	v = GetIntSliceValue("intSlice", defaultValue)

	Assert(t, []int{1, 2}, Equal(v))

	//error type
	v = GetIntSliceValue("float", defaultValue)

	Assert(t, defaultValue, Equal(v))
}

func TestGetStringSliceValue(t *testing.T) {
	createMockApolloConfig(120)
	defaultValue := []string{"100"}

	//test default
	v := GetStringSliceValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

	//normal value
	v = GetStringSliceValue("stringSlice", defaultValue)

	Assert(t, []string{"1", "2"}, Equal(v))

	//error type
	v = GetStringSliceValue("float", defaultValue)

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

func TestAutoSyncConfigServicesNormal2NotModified(t *testing.T) {
	server := runLongNotmodifiedConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.IP = server.URL
	time.Sleep(1 * time.Second)
	appConfig := env.GetPlainAppConfig()
	appConfig.NextTryConnTime = 0

	notify.AutoSyncConfigServicesSuccessCallBack([]byte(configResponseStr))

	config := env.GetCurrentApolloConfig()[newAppConfig.NamespaceName]

	fmt.Println("sleeping 10s")

	time.Sleep(10 * time.Second)

	fmt.Println("checking agcache time left")
	defaultConfigCache := GetDefaultConfigCache()

	defaultConfigCache.Range(func(key, value interface{}) bool {
		Assert(t, value, NotNilVal())
		return true
	})

	Assert(t, "100004458", Equal(config.AppID))
	Assert(t, "default", Equal(config.Cluster))
	Assert(t, testDefaultNamespace, Equal(config.NamespaceName))
	Assert(t, "20170430092936-dee2d58e74515ff3", Equal(config.ReleaseKey))
	Assert(t, "value1", Equal(GetStringValue("key1", "")))
	Assert(t, "value2", Equal(GetStringValue("key2", "")))
	checkBackupFile(t)
}

func checkBackupFile(t *testing.T) {
	appConfig := env.GetPlainAppConfig()
	newConfig, e := extension.GetFileHandler().LoadConfigFile(appConfig.GetBackupConfigPath(), testDefaultNamespace)
	t.Log(newConfig.Configurations)
	Assert(t, e, NilVal())
	Assert(t, newConfig.Configurations, NotNilVal())
	for k, v := range newConfig.Configurations {
		Assert(t, GetStringValue(k, ""), Equal(v))
	}
}

func runLongNotmodifiedConfigResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Microsecond)
		w.WriteHeader(http.StatusNotModified)
	}))

	return ts
}

func TestConfig_GetStringValue(t *testing.T) {
	createMockApolloConfig(120)
	config := GetConfig(testDefaultNamespace)

	defaultValue := "j"
	//test default
	v := config.GetStringValue("joe", defaultValue)
	Assert(t, defaultValue, Equal(v))

	//normal value
	v = config.GetStringValue("string", defaultValue)

	Assert(t, "value", Equal(v))
}

func TestConfig_GetBoolValue(t *testing.T) {
	createMockApolloConfig(120)
	defaultValue := false
	config := GetConfig(testDefaultNamespace)

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
	createMockApolloConfig(120)
	defaultValue := 100000.1
	config := GetConfig(testDefaultNamespace)

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
	createMockApolloConfig(120)
	defaultValue := 100000
	config := GetConfig(testDefaultNamespace)

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

func TestGetApolloConfigCache(t *testing.T) {
	cache := GetApolloConfigCache()
	Assert(t, cache, NotNilVal())
}
