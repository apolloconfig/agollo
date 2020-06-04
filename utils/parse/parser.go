package parse

import (
	"github.com/zouyx/agollo/v3/agcache"
)

//ContentParser 内容转换
type ContentParser interface {
	Parse(cache agcache.CacheInterface) (string, error)
}
