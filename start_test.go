package agollo

import (
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	Start()

	time.Sleep(2*time.Second)
}
