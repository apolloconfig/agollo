package jsonconfig

import (
	"testing"
	"github.com/zouyx/agollo/test"
)

func TestLoad(t *testing.T) {
	con:=Load()
	t.Log(con)
	test.NotNil(t,con)



}