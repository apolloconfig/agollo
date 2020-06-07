package normal

import (
	"testing"

	. "github.com/tevid/gohamcrest"
)

var (
	defaultParser = &Parser{}
)

func TestDefaultParser(t *testing.T) {
	s, err := defaultParser.Parse(`aaaa`)
	Assert(t, err, NilVal())
	Assert(t, s, NilVal())
}
