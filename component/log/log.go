// Copyright 2025 Apollo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

// Logger is the global logger instance used throughout the application
var Logger LoggerInterface

func init() {
	Logger = &DefaultLogger{}
}

// InitLogger initializes the logger with a custom implementation
// Parameters:
//   - ILogger: The logger implementation to be used
func InitLogger(ILogger LoggerInterface) {
	Logger = ILogger
}

// LoggerInterface defines the contract for logger implementations
// This interface provides methods for different logging levels and formats
type LoggerInterface interface {
	// Debugf logs debug messages with formatting
	Debugf(format string, params ...interface{})

	// Infof logs information messages with formatting
	Infof(format string, params ...interface{})

	// Warnf logs warning messages with formatting
	Warnf(format string, params ...interface{})

	// Errorf logs error messages with formatting
	Errorf(format string, params ...interface{})

	// Debug logs debug messages
	Debug(v ...interface{})

	// Info logs information messages
	Info(v ...interface{})

	// Warn logs warning messages
	Warn(v ...interface{})

	// Error logs error messages
	Error(v ...interface{})
}

// Debugf formats and logs a debug message using the global logger
// Parameters:
//   - format: The format string
//   - params: The parameters to be formatted
func Debugf(format string, params ...interface{}) {
	Logger.Debugf(format, params...)
}

// Infof formats and logs an information message using the global logger
// Parameters:
//   - format: The format string
//   - params: The parameters to be formatted
func Infof(format string, params ...interface{}) {
	Logger.Infof(format, params...)
}

// Warnf formats and logs a warning message using the global logger
// Parameters:
//   - format: The format string
//   - params: The parameters to be formatted
func Warnf(format string, params ...interface{}) {
	Logger.Warnf(format, params...)
}

// Errorf formats and logs an error message using the global logger
// Parameters:
//   - format: The format string
//   - params: The parameters to be formatted
func Errorf(format string, params ...interface{}) {
	Logger.Errorf(format, params...)
}

// Debug logs a debug message using the global logger
// Parameters:
//   - v: The values to be logged
func Debug(v ...interface{}) {
	Logger.Debug(v...)
}

// Info logs an information message using the global logger
// Parameters:
//   - v: The values to be logged
func Info(v ...interface{}) {
	Logger.Info(v...)
}

// Warn logs a warning message using the global logger
// Parameters:
//   - v: The values to be logged
func Warn(v ...interface{}) {
	Logger.Warn(v...)
}

// Error logs an error message using the global logger
// Parameters:
//   - v: The values to be logged
func Error(v ...interface{}) {
	Logger.Error(v...)
}

// DefaultLogger provides a basic implementation of the LoggerInterface
// This implementation is used as a fallback when no custom logger is provided
type DefaultLogger struct {
}

// Debugf implements debug level formatted logging (empty implementation)
func (d *DefaultLogger) Debugf(format string, params ...interface{}) {
}

// Infof implements info level formatted logging (empty implementation)
func (d *DefaultLogger) Infof(format string, params ...interface{}) {
}

// Warnf implements warning level formatted logging (empty implementation)
func (d *DefaultLogger) Warnf(format string, params ...interface{}) {
}

// Errorf implements error level formatted logging (empty implementation)
func (d *DefaultLogger) Errorf(format string, params ...interface{}) {
}

// Debug implements debug level logging (empty implementation)
func (d *DefaultLogger) Debug(v ...interface{}) {
}

// Info implements info level logging (empty implementation)
func (d *DefaultLogger) Info(v ...interface{}) {
}

// Warn implements warning level logging (empty implementation)
func (d *DefaultLogger) Warn(v ...interface{}) {
}

// Error implements error level logging (empty implementation)
func (d *DefaultLogger) Error(v ...interface{}) {
}
