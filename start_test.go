package agollo

import (
	"testing"
	"github.com/zouyx/agollo/test"
)

func TestStart(t *testing.T) {
	go runMockConfigServer(onlyNormalConfigResponse)
	go runMockNotifyServer(onlyNormalResponse)
	defer closeMockConfigServer()

	Start()

	value := getValue("key1")
	test.Equal(t,"value1",value)
}
