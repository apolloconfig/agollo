package agollo

import (
	. "github.com/tevid/gohamcrest"
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	go runMockConfigServer(onlyNormalConfigResponse)
	go runMockNotifyServer(onlyNormalResponse)
	defer closeMockConfigServer()

	Start()

	value := getValue("key1")
	Assert(t, "value1", Equal(value))
}

func TestErrorStart(t *testing.T) {
	server := runErrorResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	time.Sleep(1 * time.Second)

	Start()

	value := getValue("key1")
	Assert(t, "value1", Equal(value))

	value2 := getValue("key2")
	Assert(t, "value2", Equal(value2))

}
