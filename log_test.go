package agollo

import (
	"github.com/cihub/seelog"
	"testing"
)

func TestInitNullSeeLog(t *testing.T) {
	initSeeLog("joe.config")

	seelog.Error("good girl!")
}
