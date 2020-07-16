package conver

import (
	. "github.com/tevid/gohamcrest"
	"testing"
)

func TestConverterJSON(t *testing.T) {
	c := NewConverter("json")

	testContent1 := `{"test":1}`
	result, err := c.ConvertToMap(testContent1)
	Assert(t, err, NilVal())
	Assert(t, result, NotNilVal())
	Assert(t, result["test"], Equal(float64(1)))

	result, err = c.ConvertToMap("")
	Assert(t, err, NotNilVal())
}

func TestConverterYAML(t *testing.T) {
	c := NewConverter("yaml")

	testContent1 := `
a:
    a1: a1
`
	result, err := c.ConvertToMap(testContent1)
	Assert(t, err, NilVal())
	Assert(t, result, NotNilVal())
	Assert(t, result["a.a1"], Equal("a1"))
}
