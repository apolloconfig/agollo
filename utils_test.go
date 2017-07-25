package agollo

import (
	"testing"
	"strings"
	"github.com/zouyx/agollo/test"
)

func TestGetInternal(t *testing.T) {
	//fmt.Println("Usage of ./getmyip --get_ip=(external|internal)")
	//flag.Parse()
	ip:=GetInternal()

	t.Log("Internal ip:",ip)
	nums:=strings.Split(ip,".")

	test.Equal(t,4,len(nums))
}

func TestIsEmpty(t *testing.T) {
	flag:=IsEmpty("")
	test.Equal(t,true,flag)

	flag=IsEmpty("abc")
	test.Equal(t,false,flag)
}

func TestIsNotEmpty(t *testing.T) {
	flag:=IsNotEmpty("")
	test.Equal(t,false,flag)

	flag=IsNotEmpty("abc")
	test.Equal(t,true,flag)
}

func TestIsNil(t *testing.T) {
	flag:=IsNil(nil)
	test.Equal(t,true,flag)

	flag=IsNil("")
	test.Equal(t,false,flag)
}

func TestIsNotNil(t *testing.T) {
	flag:=IsNotNil(nil)
	test.Equal(t,false,flag)

	flag=IsNotNil("")
	test.Equal(t,true,flag)
}