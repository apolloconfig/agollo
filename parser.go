package agollo

import (
	"fmt"
	"github.com/zouyx/agollo/agcache"
)

const propertiesFormat ="%s=%s\n"

type ContentParser interface {
	parse(cache agcache.CacheInterface) (string,error)
}

type DefaultParser struct {

}

func (d *DefaultParser)parse(cache agcache.CacheInterface) (string,error){
	value, err := cache.Get(defaultContentKey)
	if err!=nil{
		return "",err
	}
	return string(value),nil
}

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