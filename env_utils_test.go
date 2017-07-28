package agollo

import (
	"testing"
	"fmt"
	"github.com/zouyx/agollo/test"
)

func createEnvMap()map[string]env {
	envMap:=make(map[string]env)
	envMap["LOCAL"]=local
	envMap["local"]=local
	envMap["DEV"]=dev
	envMap["dev"]=dev
	envMap["FWS"]=fws
	envMap["fws"]=fws
	envMap["FAT"]=fat
	envMap["fat"]=fat
	envMap["UAT"]=uat
	envMap["uat"]=uat
	envMap["LPT"]=lpt
	envMap["lpt"]=lpt
	envMap["PRO"]=pro
	envMap["pro"]=pro
	envMap["PROD"]=pro
	envMap["prod"]=pro
	envMap["TOOLS"]=tools
	envMap["tools"]=tools
	envMap["213123"]=unknown
	envMap["jjj"]=unknown
	envMap[""]=unknown

	return envMap
}

func TestTransformEnv(t *testing.T) {

	envMap:=createEnvMap()

	for key,value:=range envMap{
		env:=transformEnv(key)
		t.Log(fmt.Sprintf("对比:%s,期望值:%d,实际值:%d",key,+value,env))
		test.Equal(t,value,env)
	}

}
