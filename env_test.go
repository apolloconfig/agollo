package agollo

import (
	"fmt"
	. "github.com/tevid/gohamcrest"
	"testing"
)

func TestFromString(t *testing.T) {

	envMap := createEnvMap()

	for key, value := range envMap {
		environment, err := fromString(key)
		t.Log(fmt.Sprintf("对比:%s,期望值:%d,实际值:%d", key, value, environment))
		if unknown == environment {
			Assert(t, err,NotNilVal())
		} else {
			Assert(t, err,NilVal())
		}
		Assert(t, value, Equal(environment))
	}

}
