package agollo

import (
	"net"
	"os"
	"reflect"
)

var(
	internalIp string
)

//ips
func getInternal() string {
	if internalIp!=""{
		return internalIp
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops:" + err.Error())
		os.Exit(1)
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				internalIp=ipnet.IP.To4().String()
				return internalIp
			}
		}
	}
	return ""
}


//stringUtils
func isEmpty(str string) bool {
	return ""==str
}

func isNotEmpty(str string) bool {
	return !isEmpty(str)
}

//objectUtils
func isNil(object interface{}) bool {

	return isNilObject(object)
}

func isNotNil(object interface{}) bool {
	return !isNilObject(object)
}


func isNilObject(object interface{}) bool {
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
