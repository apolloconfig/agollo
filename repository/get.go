package repository

import (
	"strconv"
	"github.com/cihub/seelog"
)

func getValue(key string)string{
	value:=getConfigValue(key)
	if value==nil{
		return empty
	}

	return value.(string)
}

func GetStringValue(key string,defaultValue string)string{
	value:=getValue(key)
	if value==empty{
		return defaultValue
	}

	return value
}

func GetIntValue(key string,defaultValue int) int {
	value :=getValue(key)

	i,err:=strconv.Atoi(value)
	if err!=nil{
		seelog.Error("convert to int fail!error:",err)
		return defaultValue
	}

	return i
}

func GetFloatValue(key string,defaultValue float64) float64 {
	value :=getValue(key)

	i,err:=strconv.ParseFloat(value,64)
	if err!=nil{
		seelog.Error("convert to float fail!error:",err)
		return defaultValue
	}

	return i
}

func GetBoolValue(key string,defaultValue bool) bool {
	value :=getValue(key)

	b,err:=strconv.ParseBool(value)
	if err!=nil{
		seelog.Error("convert to bool fail!error:",err)
		return defaultValue
	}

	return b
}
