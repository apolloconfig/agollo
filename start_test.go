package agollo

import (
	"testing"
	"github.com/zouyx/agollo/test"
	"time"
)

func TestStart(t *testing.T) {
	go runMockConfigServer(onlyNormalConfigResponse)
	go runMockNotifyServer(onlyNormalResponse)
	defer closeMockConfigServer()

	Start()

	value := getValue("key1")
	test.Equal(t,"value1",value)
}

func TestErrorStart(t *testing.T) {
	server:= runErrorResponse()
	newAppConfig:=getTestAppConfig()
	newAppConfig.Ip=server.URL

	time.Sleep(1 * time.Second)

	Start()

	value := getValue("key1")
	test.Equal(t,"value1",value)

	value2 := getValue("key2")
	test.Equal(t,"value2",value2)

}