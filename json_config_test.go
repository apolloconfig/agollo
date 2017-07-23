package agollo

import (
	"testing"
)

func TestLoadJsonConfig(t *testing.T) {
	con:=LoadJsonConfig()
	t.Log(con)
	NotNil(t,con)
	Equal(t,"soa_recommend_shunt",con.AppId)
	Equal(t,"dev",con.Cluster)
	Equal(t,"application",con.NamespaceName)

}