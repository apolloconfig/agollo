package dto

import (
	"testing"
	"encoding/json"
)

func TestCreateApolloConfig(t *testing.T) {
	c:=CreateApolloConfig("soa_recommend_shunt","dev","application","")
	b,e:=json.Marshal(c)
	t.Log(e)
	t.Log(c)
	t.Log(b)
	t.Log(string(b))

}
