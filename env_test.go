package agollo

import (
	"testing"
	"fmt"
)

func TestFromString(t *testing.T) {

	envMap := createEnvMap()

	for key, value := range envMap {
		env, err := FromString(key)
		t.Log(fmt.Sprintf("对比:%s,期望值:%d,实际值:%d", key, +value, env))
		if (UNKNOWN == env) {
			NotNil(t, err)
		} else {
			Nil(t, err)
		}
		Equal(t, value, env)
	}

}
