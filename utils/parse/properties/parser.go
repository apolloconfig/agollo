package properties

import (
	"fmt"
	"github.com/zouyx/agollo/v3/agcache"
	"github.com/zouyx/agollo/v3/constant"
	"github.com/zouyx/agollo/v3/extension"
	"github.com/zouyx/agollo/v3/utils"
)

func init() {
	extension.AddFormatParser(constant.Properties, &Parser{})
}

const (
	propertiesFormat = "%s=%s\n"
)

// Parser properties转换器
type Parser struct {
}

// Parse 内存内容=>properties文件转换器
func (d *Parser) Parse(cache agcache.CacheInterface) (string, error) {
	properties := convertToProperties(cache)
	return properties, nil
}

func convertToProperties(cache agcache.CacheInterface) string {
	properties := utils.Empty
	if cache == nil {
		return properties
	}
	cache.Range(func(key, value interface{}) bool {
		properties += fmt.Sprintf(propertiesFormat, key, string(value.([]byte)))
		return true
	})
	return properties
}
