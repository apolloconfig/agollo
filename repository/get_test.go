package repository

import "testing"

func TestGetStringValue(t *testing.T) {
	v:=GetStringValue("joe","j")
	t.Log(v)
}
