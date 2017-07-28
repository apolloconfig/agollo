package agollo

import (
	"testing"
	"fmt"
	"github.com/zouyx/agollo/test"
)

func TestFromString(t *testing.T) {

	envMap := createEnvMap()

	for key, value := range envMap {
		environment, err := fromString(key)
		t.Log(fmt.Sprintf("对比:%s,期望值:%d,实际值:%d", key, value, environment))
		if (unknown == environment) {
			test.NotNil(t, err)
		} else {
			test.Nil(t, err)
		}
		test.Equal(t, value, environment)
	}

}
