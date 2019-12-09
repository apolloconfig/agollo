package agollo

import (
	"fmt"
	"github.com/zouyx/agollo/v2/agcache"
)

const propertiesFormat ="%s=%s\n"

//ContentParser 内容转换
type ContentParser interface {
	parse(cache agcache.CacheInterface) (string,error)
}

//DefaultParser 默认内容转换器
type DefaultParser struct {

}

func (d *DefaultParser)parse(cache agcache.CacheInterface) (string,error){
	value, err := cache.Get(defaultContentKey)
	if err!=nil{
		return "",err
	}
	return string(value),nil
}

//PropertiesParser properties转换器
type PropertiesParser struct {

}

func (d *PropertiesParser)parse(cache agcache.CacheInterface) (string,error){
	properties := convertToProperties(cache)
	return properties,nil
}

func convertToProperties(cache agcache.CacheInterface) string {
	properties:=""
	if cache==nil {
		return properties
	}
	cache.Range(func(key, value interface{}) bool {
		properties+=fmt.Sprintf(propertiesFormat,key,string(value.([]byte)))
		return true
	})
	return properties
}