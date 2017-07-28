package log

import (
	"github.com/cihub/seelog"
	"errors"
)

func init(){
	initSeeLog("seelog.xml")
}

func initSeeLog(configPath string)  {
	logger, err := seelog.LoggerFromConfigAsFile(configPath)

	if err != nil {
		errors.New("init log fail,error!"+err.Error())
		return
	}

	logger.SetAdditionalStackDepth(1)
	seelog.ReplaceLogger(logger)
	defer seelog.Flush()
}

// Call seelog.Debugf
func Debugf(format string, params ...interface{}) {
	if seelog.Current==nil{
		return
	}
	seelog.Debugf(format,params)
}

// Call seelog.Debug
func Debug(v ...interface{}) {
	if seelog.Current==nil{
		return
	}
	seelog.Debug(v)
}

// Call seelog.Error
func Error(v ...interface{}){
	if seelog.Current==nil{
		return
	}
	seelog.Error(v)
}

// Call seelog.Warn
func Warn(v ...interface{}) error {
	if seelog.Current==nil{
		return errors.New("seelog has not init.")
	}
	return seelog.Warn(v)
}

// Call seelog.Errorf
func Errorf(format string, params ...interface{}) error {
	if seelog.Current==nil{
		return errors.New("seelog has not init.")
	}
	return seelog.Errorf(format,params)
}