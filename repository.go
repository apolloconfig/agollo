package agollo

import (
	"strconv"
	"github.com/cihub/seelog"
	"github.com/coocood/freecache"
)

const (
	empty  = ""

	//50m
	apolloConfigCacheSize=50*1024*1024

	//1 minute
	configCacheExpireTime=120
)
var (
	currentConnApolloConfig *ApolloConnConfig=&ApolloConnConfig{}

	//config from apollo
	apolloConfigCache *freecache.Cache = freecache.NewCache(apolloConfigCacheSize)
)

func updateApolloConfig(apolloConfig *ApolloConfig)  {
	if apolloConfig==nil{
		seelog.Error("apolloConfig is null,can't update!")
		return
	}
	go updateApolloConfigCache(apolloConfig.Configurations,configCacheExpireTime)

	//update apollo connection config

	currentConnApolloConfig.Lock()
	defer currentConnApolloConfig.Unlock()
	currentConnApolloConfig=&apolloConfig.ApolloConnConfig
}

func updateApolloConfigCache(configurations map[string]string,expireTime int)  {
	if configurations==nil||len(configurations)==0{
		return
	}

	//get old keys
	mp := map[string]bool{}
	it := apolloConfigCache.NewIterator()
	for en := it.Next(); en != nil; en = it.Next() {
		mp[string(en.Key)] = true
	}

	// update new keys
	for key, value := range configurations {
		apolloConfigCache.Set([]byte(key), []byte(value), expireTime)
		delete(mp, string(key))
	}

	// remove del keys
	for key := range mp {
		apolloConfigCache.Del([]byte(key))
	}
}

func touchApolloConfigCache() error{
	updateApolloConfigCacheTime(configCacheExpireTime)
	return nil
}

func updateApolloConfigCacheTime(expireTime int)  {
	it := apolloConfigCache.NewIterator()
	for i := int64(0); i < apolloConfigCache.EntryCount(); i++ {
		entry := it.Next()
		if entry==nil{
			break
		}
		apolloConfigCache.Set([]byte(entry.Key),[]byte(entry.Value),expireTime)
	}
}

func GetApolloConfigCache() *freecache.Cache {
	return apolloConfigCache
}

func GetCurrentApolloConfig()*ApolloConnConfig  {
	currentConnApolloConfig.RLock()
	defer currentConnApolloConfig.RUnlock()
	return currentConnApolloConfig
}

func getConfigValue(key string) interface{}  {
	value,err:=apolloConfigCache.Get([]byte(key))
	if err!=nil{
		seelog.Error("get config value fail!err:",err)
		return empty
	}

	return string(value)
}


func getValue(key string)string{
	value:=getConfigValue(key)
	if value==nil{
		return empty
	}

	return value.(string)
}

func GetStringValue(key string,defaultValue string)string{
	value:=getValue(key)
	if value==empty{
		return defaultValue
	}

	return value
}

func GetIntValue(key string,defaultValue int) int {
	value :=getValue(key)

	i,err:=strconv.Atoi(value)
	if err!=nil{
		seelog.Debug("convert to int fail!error:",err)
		return defaultValue
	}

	return i
}

func GetFloatValue(key string,defaultValue float64) float64 {
	value :=getValue(key)

	i,err:=strconv.ParseFloat(value,64)
	if err!=nil{
		seelog.Debug("convert to float fail!error:",err)
		return defaultValue
	}

	return i
}

func GetBoolValue(key string,defaultValue bool) bool {
	value :=getValue(key)

	b,err:=strconv.ParseBool(value)
	if err!=nil{
		seelog.Debug("convert to bool fail!error:",err)
		return defaultValue
	}

	return b
}
