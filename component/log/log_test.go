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
