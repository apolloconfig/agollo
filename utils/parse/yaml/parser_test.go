package yaml

import (
	"github.com/zouyx/agollo/v3/utils"
	"github.com/zouyx/agollo/v3/utils/parse"
	"testing"

	. "github.com/tevid/gohamcrest"
)

var (
	yamlParser parse.ContentParser = &Parser{}
)

func TestYAMLParser(t *testing.T) {
	s, err := yamlParser.Parse(`
a:
    a1: a1
b:
    b1: b1
c:
    c1: c1
d:
    d1: d1
e:  
    e1: e1`)
	Assert(t, err, NilVal())

	Assert(t, s["a.a1"], Equal("a1"))

	Assert(t, s["b.b1"], Equal("b1"))

	Assert(t, s["c.c1"], Equal("c1"))

}

func TestYAMLParserOnException(t *testing.T) {
	s, err := yamlParser.Parse(utils.Empty)
	Assert(t, err, NilVal())
	Assert(t, s, NilVal())
	s, err = yamlParser.Parse(0)
	Assert(t, err, NilVal())
	Assert(t, s, NilVal())

	m := convertToMap(nil)
	Assert(t, m, NilVal())
}

