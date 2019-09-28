package agollo

var logger LoggerInterface

func init() {
	logger=&DefaultLogger{}
}

func initLogger(ILogger LoggerInterface) {
	logger = ILogger
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

type DefaultLogger struct {
}

func (this *DefaultLogger)Debugf(format string, params ...interface{})  {
	
}

func (this *DefaultLogger)Infof(format string, params ...interface{}) {

}


func (this *DefaultLogger)Warnf(format string, params ...interface{}) error {
	return nil
}

func (this *DefaultLogger)Errorf(format string, params ...interface{}) error {
	return nil
}


func (this *DefaultLogger)Debug(v ...interface{}) {

}
func (this *DefaultLogger)Info(v ...interface{}){

}

func (this *DefaultLogger)Warn(v ...interface{}) error{
	return nil
}

func (this *DefaultLogger)Error(v ...interface{}) error{
	return nil
}