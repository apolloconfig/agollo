package log

import "testing"

func TestDebugf(t *testing.T) {
	Logger.Debugf("")
}

func TestInfof(t *testing.T) {
	Logger.Infof("")
}

func TestErrorf(t *testing.T) {
	Logger.Errorf("")
}

func TestWarnf(t *testing.T) {
	Logger.Warnf("")
}

func TestDebug(t *testing.T) {
	Logger.Debug("")
}

func TestInfo(t *testing.T) {
	Logger.Info("")
}

func TestError(t *testing.T) {
	Logger.Error("")
}

func TestWarn(t *testing.T) {
	Logger.Warn("")
}

func TestInitLogger(t *testing.T) {
	InitLogger(Logger)
}

func TestCommonDebugf(t *testing.T) {
	Debugf("")
}

func TestCommonInfof(t *testing.T) {
	Infof("")
}

func TestCommonErrorf(t *testing.T) {
	Errorf("")
}

func TestCommonWarnf(t *testing.T) {
	Warnf("")
}

func TestCommonDebug(t *testing.T) {
	Debug("")
}

func TestCommonInfo(t *testing.T) {
	Info("")
}

func TestCommonError(t *testing.T) {
	Error("")
}

func TestCommonWarn(t *testing.T) {
	Warn("")
}
