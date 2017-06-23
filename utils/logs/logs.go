package logs

import (
	"github.com/cihub/seelog"
)

func init(){
	logger, err := seelog.LoggerFromConfigAsFile("seelog.xml")

	if err != nil {
		panic("init log fail,error!"+err.Error())
	}

	logger.SetAdditionalStackDepth(1)
	seelog.ReplaceLogger(logger)
	defer seelog.Flush()
}

