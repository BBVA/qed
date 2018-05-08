// copyright © 2018 banco bilbao vizcaya argentaria s.a.  all rights reserved.
// use of this source code is governed by an apache 2 license
// that can be found in the license file

package log

import (
	"fmt"
	"log"
	"os"
)

// Log levels constants
const (
	SILENT = "silent"
	ERROR  = "error"
	INFO   = "info"
	DEBUG  = "debug"
)

// Private interface for the std variable
type logger interface {
	Error(v ...interface{})
	Errorf(format string, v ...interface{})

	Info(v ...interface{})
	Infof(format string, v ...interface{})

	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
}

// The default logger is an log.ERROR level.
var std logger = newError(os.Stdout, "Qed: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)

// Below is the public interface for the logger, a proxy for the switchable
// implementation defined in std

func Error(v ...interface{}) {
	std.Error(v...)
}

var (
	Fatal func(...interface{}) = Error
	Panic func(...interface{}) = Error
)

func Errorf(format string, v ...interface{}) {
	std.Errorf(format, v...)
}

var (
	Fatalf func(string, ...interface{}) = Errorf
	Panicf func(string, ...interface{}) = Errorf
)

func Info(v ...interface{}) {
	std.Info(v...)
}

func Infof(format string, v ...interface{}) {
	std.Infof(format, v...)
}

func Debug(v ...interface{}) {
	std.Debug(v...)
}

func Debugf(format string, v ...interface{}) {
	std.Debugf(format, v...)
}

func SetLogger(namespace, level string) {

	prefix := fmt.Sprintf("%s: ", namespace)

	switch level {
	case SILENT:
		std = newSilent()
	case ERROR:
		std = newError(os.Stdout, prefix, log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
	case INFO:
		std = newInfo(os.Stdout, prefix, log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
	case DEBUG:
		std = newDebug(os.Stdout, prefix, log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
	default:
		l := newInfo(os.Stdout, prefix, log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
		l.Infof("Incorrect level of verbosity (%d) fallback to log.INFO", level)
		std = l
	}

}
