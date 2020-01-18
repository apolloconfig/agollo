package utils

import (
	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v3/agcache"
	"strings"
	"testing"
)

var (
	testDefaultCache agcache.CacheInterface
	defaultParser    ContentParser
	propertiesParser ContentParser
)

func init() {
	factory := &agcache.DefaultCacheFactory{}
	testDefaultCache = factory.Create()

	defaultParser = &DefaultParser{}

	propertiesParser = &PropertiesParser{}

	testDefaultCache.Set("a", []byte("b"), 100)
	testDefaultCache.Set("c", []byte("d"), 100)
	testDefaultCache.Set("content", []byte("content"), 100)
}

func TestDefaultParser(t *testing.T) {
	s, err := defaultParser.Parse(testDefaultCache)
	Assert(t, err, NilVal())
	Assert(t, s, Equal("content"))

	s, err = defaultParser.Parse(nil)
	Assert(t, err, NilVal())
	Assert(t, s, Equal(Empty))
}

func TestPropertiesParser(t *testing.T) {
	s, err := propertiesParser.Parse(testDefaultCache)
	Assert(t, err, NilVal())

	hasString := strings.Contains(s, "a=b")
	Assert(t, hasString, Equal(true))

	hasString = strings.Contains(s, "c=d")
	Assert(t, hasString, Equal(true))

	hasString = strings.Contains(s, "content=content")
	Assert(t, hasString, Equal(true))

	s, err = defaultParser.Parse(nil)
	Assert(t, err, NilVal())
	Assert(t, s, Equal(Empty))
}
