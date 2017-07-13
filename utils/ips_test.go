package utils

import (
	"testing"
)

func TestGetInternal(t *testing.T) {
	//fmt.Println("Usage of ./getmyip --get_ip=(external|internal)")
	//flag.Parse()
	t.Log(GetInternal())
}