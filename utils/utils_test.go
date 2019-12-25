package utils

import (
	. "github.com/tevid/gohamcrest"
	"strings"
	"testing"
)

func TestGetInternal(t *testing.T) {
	//fmt.Println("Usage of ./getmyip --get_ip=(external|internal)")
	//flag.Parse()
	ip := GetInternal()

	t.Log("Internal ip:", ip)
	nums := strings.Split(ip, ".")

	Assert(t, true, Equal(len(nums) > 0))
}

func TestIsNotNil(t *testing.T) {
	flag := IsNotNil(nil)
	Assert(t, false, Equal(flag))

	flag = IsNotNil("")
	Assert(t, true, Equal(flag))
}
