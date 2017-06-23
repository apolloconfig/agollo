package config

import (
	"testing"
	"os"
)

func TestInitRefreshInterval(t *testing.T) {
	os.Setenv(REFRESH_INTERVAL_KEY,"joe")
	initRefreshInterval()
}