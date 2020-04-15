package extension

import (
	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v3/env/file"
	"testing"
)

func TestInitFileHandler(t *testing.T) {
	InitFileHandler()
	Assert(t, file.GetFileHandler(), NotNilVal())
}
