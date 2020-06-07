package utils

import (
	"net"
	"os"
	"reflect"
	"sync"
)

const (
	//Empty 空字符串
	Empty = ""
)

var (
	internalIPOnce sync.Once
	internalIP     = ""

	// EmptyMap 空map
	EmptyMap = make(map[string]string)
)

//GetInternal 获取内部ip
func GetInternal() string {
	internalIPOnce.Do(func() {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			os.Stderr.WriteString("Oops:" + err.Error())
			os.Exit(1)
		}
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					internalIP = ipnet.IP.To4().String()
				}
			}
		}
	})
	return internalIP
}

//IsNotNil 判断是否nil
func IsNotNil(object interface{}) bool {
	return !IsNilObject(object)
}

//IsNilObject 判断是否空对象
func IsNilObject(object interface{}) bool {
	if object == nil {
		return true
	}

	value := reflect.ValueOf(object)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return true
	}

	return false
}
