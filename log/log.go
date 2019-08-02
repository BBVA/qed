/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package log implements the qed/log wrapper that formats the logs in our
// custom format as well as logging levels.
package log

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/logutils"
)

// Log levels constants
const (
	SILENT = "silent"
	ERROR  = "error"
	INFO   = "info"
	DEBUG  = "debug"

	caller = 3
)

// Private interface for the std variable.
type logger interface {
	Error(v ...interface{})
	Errorf(format string, v ...interface{})

	Info(v ...interface{})
	Infof(format string, v ...interface{})

	Debug(v ...interface{})
	Debugf(format string, v ...interface{})

	GetLogger() *log.Logger
	GetLoggerLevel() string
}

func getFilter(lv string) *logutils.LevelFilter {

	mapLevel := map[string]logutils.LogLevel{
		ERROR: "ERROR",
		INFO:  "INFO",
		DEBUG: "DEBUG",
	}

	return &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: mapLevel[lv],
		Writer:   os.Stdout,
	}
}

// The default logger is an log.ERROR level.
var std logger = newError(getFilter(ERROR), "Qed: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)

// To allow mocking we require a switchable variable.
var osExit = os.Exit

// Below is the public interface for the logger, a proxy for the switchable
// implementation defined in std.

// Error is the public log function to write to stdOut and stop execution.
func Error(v ...interface{}) {
	std.Error(v...)
}

var (

	// Fatal is the public log function to write to stdOut and stop execution
	// Same as Error.
	Fatal = Error

	// Panic is the public log function to write to stdOut and stop execution
	// Same as Error.
	Panic = Error
)

// Errorf is the public log function with params to write to stdOut and stop
// execution.
func Errorf(format string, v ...interface{}) {
	std.Errorf(format, v...)
}

var (

	// Fatalf is the public log function with params to write to stdOut and
	// stop execution. Same as Errorf
	Fatalf = Errorf

	// Panicf is the public log function with params to write to stdOut and
	// stop execution. Same as Errorf
	Panicf = Errorf
)

// Info is the public log function to write information relative to the usage
// of the qed package.
func Info(v ...interface{}) {
	std.Info(v...)
}

// Info is the public log function to write information with params relative
// to the usage of the qed package.
func Infof(format string, v ...interface{}) {
	std.Infof(format, v...)
}

// Debug is the public log function to write information relative to internal
// debug information.
func Debug(v ...interface{}) {
	std.Debug(v...)
}

// Debugf is the public log function to write information with params relative
// to internal debug information.
func Debugf(format string, v ...interface{}) {
	std.Debugf(format, v...)
}

// GetLogger returns a default log.Logger instance. Useful to let third party
// modules to use the same formatting options that the defined here.
func GetLogger() *log.Logger {
	return std.GetLogger()
}

// GetLoggerLevel returns the string representation of the log.Logger level.
func GetLoggerLevel() string {
	return std.GetLoggerLevel()
}

// SetLogger is a function that switches between verbosity loggers. Default
// is error level. Available levels are "silent", "debug", "info" and "error".
func SetLogger(namespace, lv string) {

	prefix := fmt.Sprintf("%s ", namespace)

	switch lv {
	case SILENT:
		std = newSilent()
	case ERROR:
		std = newError(getFilter(lv), prefix, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	case INFO:
		std = newInfo(getFilter(lv), prefix, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	case DEBUG:
		std = newDebug(getFilter(lv), prefix, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	default:
		l := newInfo(getFilter(INFO), prefix, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
		l.Infof("Incorrect level of verbosity (%v) fallback to log.INFO", lv)
		std = l
	}

}
