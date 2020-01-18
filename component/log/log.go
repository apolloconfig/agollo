package log

//Logger logger 对象
var Logger LoggerInterface

func init() {
	Logger = &DefaultLogger{}
}

//InitLogger 初始化logger对象
func InitLogger(ILogger LoggerInterface) {
	Logger = ILogger
}

type LoggerInterface interface {
	Debugf(format string, params ...interface{})

	Infof(format string, params ...interface{})

	Warnf(format string, params ...interface{}) error

	Errorf(format string, params ...interface{}) error

	Debug(v ...interface{})

	Info(v ...interface{})

	Warn(v ...interface{}) error

	Error(v ...interface{}) error
}

//Debugf debug 格式化
func Debugf(format string, params ...interface{}) {
	Logger.Debugf(format, params)
}

//Infof 打印info
func Infof(format string, params ...interface{}) {
	Logger.Infof(format, params)
}

//Warnf warn格式化
func Warnf(format string, params ...interface{}) error {
	return Logger.Warnf(format, params)
}

//Errorf error格式化
func Errorf(format string, params ...interface{}) error {
	return Logger.Errorf(format, params)
}

//Debug 打印debug
func Debug(v ...interface{}) {
	Logger.Debug(v)
}

//Info 打印Info
func Info(v ...interface{}) {
	Logger.Info(v)
}

//Warn 打印Warn
func Warn(v ...interface{}) {
	Logger.Warn(v)
}

//Error 打印Error
func Error(v ...interface{}) {
	Logger.Error(v)
}

type DefaultLogger struct {
}

//Debugf debug 格式化
func (this *DefaultLogger) Debugf(format string, params ...interface{}) {

}

//Infof 打印info
func (this *DefaultLogger) Infof(format string, params ...interface{}) {

}

//Warnf warn格式化
func (this *DefaultLogger) Warnf(format string, params ...interface{}) error {
	return nil
}

//Errorf error格式化
func (this *DefaultLogger) Errorf(format string, params ...interface{}) error {
	return nil
}

//Debug 打印debug
func (this *DefaultLogger) Debug(v ...interface{}) {

}

//Info 打印Info
func (this *DefaultLogger) Info(v ...interface{}) {

}

//Warn 打印Warn
func (this *DefaultLogger) Warn(v ...interface{}) error {
	return nil
}

//Error 打印Error
func (this *DefaultLogger) Error(v ...interface{}) error {
	return nil
}
