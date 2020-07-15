package json

import (
	"github.com/zouyx/agollo/v3/utils"
	"github.com/zouyx/agollo/v3/utils/parse"
	"testing"

	. "github.com/tevid/gohamcrest"
)

var (
	jsonParser parse.ContentParser = &Parser{}
)

func TestJSONParser(t *testing.T) {
	s, err := jsonParser.Parse(`
{"testValue": 1,
"testObject": {
   "K1": "a1"
},
"testList": [
"l1", 
"l2"
]}
`)
	Assert(t, err, NilVal())
	t.Logf("%+v", s)
	Assert(t, s["testvalue"], Equal(float64(1)))
	Assert(t, s["testobject.k1"], Equal("a1"))
	Assert(t, s["testlist"], Equal([]interface{}{"l1", "l2"}))


}

func TestJSONParserOnException(t *testing.T) {
	s, err := jsonParser.Parse(utils.Empty)
	Assert(t, err, NilVal())
	Assert(t, s, NilVal())
	s, err = jsonParser.Parse(0)
	Assert(t, err, NilVal())
	Assert(t, s, NilVal())

	m := convertToMap(nil)
	Assert(t, m, NilVal())
}

