package agollo

import (
	"errors"
	"fmt"
)

const (
	unknown= iota
	local
	dev
	fws
	fat
	uat
	lpt
	pro
	tools
)

//环境
type env int

func fromString(envKey string) (env,error) {
	environment := transformEnv(envKey)
	if environment==unknown{
		return environment,errors.New(fmt.Sprintf("Env %s is invalid",envKey))
	}
	return environment,nil
}
