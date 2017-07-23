package agollo

import (
	"errors"
	"fmt"
)

const (
	UNKNOWN= iota
	LOCAL
	DEV
	FWS
	FAT
	UAT
	LPT
	PRO
	TOOLS
)

//环境
type Env int

func FromString(env string) (Env,error) {
	environment := transformEnv(env)
	if environment==UNKNOWN{
		return environment,errors.New(fmt.Sprintf("Env %s is invalid",env))
	}
	return environment,nil
}
