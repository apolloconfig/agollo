package dto

import (
	"testing"
	"encoding/json"
)

func TestApolloConfig(t *testing.T) {
	config:=&ApolloConfig{
		AppId:"1",
	}
	j,err:=json.Marshal(config)

	t.Log(string(j))
	t.Log(err)
}