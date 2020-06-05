package normal

import (
	"github.com/zouyx/agollo/v3/constant"
	"github.com/zouyx/agollo/v3/extension"
	"github.com/zouyx/agollo/v3/utils"

	"github.com/zouyx/agollo/v3/agcache"
)

func init() {
	extension.AddFormatParser(constant.DEFAULT, &Parser{})
}

const (
	defaultContentKey = "content"
)

// Parser 默认内容转换器
type Parser struct {
}

// Parse 内存内容默认转换器
func (d *Parser) Parse(cache agcache.CacheInterface) (string, error) {
	if cache == nil {
		return utils.Empty, nil
	}

	value, err := cache.Get(defaultContentKey)
	if err != nil {
		return utils.Empty, err
	}
	return string(value), nil
}
