package agollo

import (
	"github.com/cihub/seelog"
)

func init(){
	initSeeLog("seelog.xml")
}

func initSeeLog(configPath string)  {
	logger, err := seelog.LoggerFromConfigAsFile(configPath)

	//if error is happen change to default config.
	if err != nil {
		logger, err = seelog.LoggerFromConfigAsBytes([]byte("<seelog />"))
	}

	logger.SetAdditionalStackDepth(1)
	seelog.ReplaceLogger(logger)
	defer seelog.Flush()
}