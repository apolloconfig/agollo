package agollo

import (
	"testing"
	"github.com/cihub/seelog"
)

func TestInitNullSeeLog(t *testing.T) {
	initSeeLog("joe.config")

	seelog.Error("good girl!")
}