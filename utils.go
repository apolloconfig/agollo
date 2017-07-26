package agollo

import (
	"net"
	"os"
	"reflect"
	"github.com/cihub/seelog"
)

var(
	internalIp string
)

//ips
func GetInternal() string {
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
func IsEmpty(str string) bool {
	return ""==str
}

func IsNotEmpty(str string) bool {
	return !IsEmpty(str)
}

//objectUtils
func IsNil(object interface{}) bool {

	return isNilObject(object)
}

func IsNotNil(object interface{}) bool {
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

//log
func init(){
	logger, err := seelog.LoggerFromConfigAsFile("seelog.xml")

	if err != nil {
		panic("init log fail,error!"+err.Error())
	}

	logger.SetAdditionalStackDepth(1)
	seelog.ReplaceLogger(logger)
	defer seelog.Flush()
}
