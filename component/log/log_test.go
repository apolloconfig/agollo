package log

import "testing"

func TestDebugf(t *testing.T) {
	logger.Debugf("")
}

func TestInfof(t *testing.T) {
	logger.Infof("")
}

func TestErrorf(t *testing.T) {
	logger.Errorf("")
}

func TestWarnf(t *testing.T) {
	logger.Warnf("")
}

func TestDebug(t *testing.T) {
	logger.Debug("")
}

func TestInfo(t *testing.T) {
	logger.Info("")
}

func TestError(t *testing.T) {
	logger.Error("")
}

func TestWarn(t *testing.T) {
	logger.Warn("")
}

func TestInitLogger(t *testing.T) {
	initLogger(logger)
}
