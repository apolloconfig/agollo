package agollo

import (
	"testing"
	"strings"
	"github.com/zouyx/agollo/test"
)

func TestGetInternal(t *testing.T) {
	//fmt.Println("Usage of ./getmyip --get_ip=(external|internal)")
	//flag.Parse()
	ip:=getInternal()

	t.Log("Internal ip:",ip)
	nums:=strings.Split(ip,".")

	test.Equal(t,4,len(nums))
}

func TestIsEmpty(t *testing.T) {
	flag:=isEmpty("")
	test.Equal(t,true,flag)

	flag=isEmpty("abc")
	test.Equal(t,false,flag)
}

func TestIsNotEmpty(t *testing.T) {
	flag:=isNotEmpty("")
	test.Equal(t,false,flag)

	flag=isNotEmpty("abc")
	test.Equal(t,true,flag)
}

func TestIsNil(t *testing.T) {
	flag:=isNil(nil)
	test.Equal(t,true,flag)

	flag=isNil("")
	test.Equal(t,false,flag)
}

func TestIsNotNil(t *testing.T) {
	flag:=isNotNil(nil)
	test.Equal(t,false,flag)

	flag=isNotNil("")
	test.Equal(t,true,flag)
}