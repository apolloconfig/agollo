package jsonconfig

import (
	"testing"
	"github.com/zouyx/agollo/test"
)

func TestLoad(t *testing.T) {
	con:=Load()
	t.Log(con)
	test.NotNil(t,con)
	test.Equal(t,"soa_recommend_shunt",con.AppId)
	test.Equal(t,"dev",con.Cluster)
	test.Equal(t,"application",con.NamespaceName)

}