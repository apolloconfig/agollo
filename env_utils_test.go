package agollo

import (
	"testing"
	"fmt"
)

func createEnvMap()map[string]Env {
	envMap:=make(map[string]Env)
	envMap["LOCAL"]=LOCAL
	envMap["local"]=LOCAL
	envMap["DEV"]=DEV
	envMap["dev"]=DEV
	envMap["FWS"]=FWS
	envMap["fws"]=FWS
	envMap["FAT"]=FAT
	envMap["fat"]=FAT
	envMap["UAT"]=UAT
	envMap["uat"]=UAT
	envMap["LPT"]=LPT
	envMap["lpt"]=LPT
	envMap["PRO"]=PRO
	envMap["pro"]=PRO
	envMap["TOOLS"]=TOOLS
	envMap["tools"]=TOOLS
	envMap["213123"]=UNKNOWN
	envMap["jjj"]=UNKNOWN

	return envMap
}

func TestTransformEnv(t *testing.T) {

	envMap:=createEnvMap()

	for key,value:=range envMap{
		env:=transformEnv(key)
		t.Log(fmt.Sprintf("对比:%s,期望值:%d,实际值:%d",key,+value,env))
		Equal(t,value,env)
	}

}
