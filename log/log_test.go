package log

import (
	"github.com/cihub/seelog"
	"testing"
)

func TestInitSeeLog(t *testing.T) {
	seelog.ReplaceLogger(nil)

	initSeeLog("joe.config")

	seelog.Error("good girl!")
}