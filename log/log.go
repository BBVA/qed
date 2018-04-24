package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

// export flags from here, so we do not need to
// import standard log every time and still
// give the same constructor signature
const (
	Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
	Ltime                         // the time in the local time zone: 01:23:23
	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	LstdFlags     = Ldate | Ltime // initial values for the standard logger
)

type Logger interface {
	Error(v ...interface{})
	Info(v ...interface{})
	Debug(v ...interface{})
	Errorf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})	
}

type NoLogger struct {
	log.Logger
}

// A impl 'l Nologger' verifiabledata/log.Logger
func (l NoLogger) Error(v ...interface{}) {return}
func (l NoLogger) Info(v ...interface{}) {return}
func (l NoLogger) Debug(v ...interface{}) {return}
func (l NoLogger) Errorf(format string, v ...interface{}) {return}
func (l NoLogger) Infof(format string, v ...interface{}) {return}
func (l NoLogger) Debugf(format string, v ...interface{}) {return}


type ErrorLogger struct {
	log.Logger
}

func NewError(out io.Writer, prefix string, flag int) *ErrorLogger {
	var l ErrorLogger

	l.SetOutput(out)
	l.SetPrefix(prefix)
	l.SetFlags(flag)
	return &l
}

// A impl 'l ErrorLogger' verifiabledata/log.Logger
func (l ErrorLogger) Error(v ...interface{}) {
	l.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

func (l ErrorLogger) Errorf(format string, v ...interface{}) {
	l.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (l ErrorLogger) Info(v ...interface{}) {return}
func (l ErrorLogger) Debug(v ...interface{}) {return}
func (l ErrorLogger) Infof(format string, v ...interface{}) {return}
func (l ErrorLogger) Debugf(format string, v ...interface{}) {return}


type InfoLogger struct {
	log.Logger
}


func NewInfo(out io.Writer, prefix string, flag int) *InfoLogger {
	var l InfoLogger

	l.SetOutput(out)
	l.SetPrefix(prefix)
	l.SetFlags(flag)
	return &l
}

// A impl 'l InfoLogger' verifiabledata/log.Logger
func (l InfoLogger) Error(v ...interface{}) {
	l.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

func (l InfoLogger) Info(v ...interface{}) {
	l.Output(2, fmt.Sprint(v...))
}

func (l InfoLogger) Errorf(format string, v ...interface{}) {
	l.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (l InfoLogger) Infof(format string, v ...interface{}) {
	l.Output(2, fmt.Sprintf(format, v...))
}

func (l InfoLogger) Debug(v ...interface{}) {return}
func (l InfoLogger) Debugf(format string, v ...interface{}) {return}


type DebugLogger struct {
	log.Logger
}

func NewDebug(out io.Writer, prefix string, flag int) *DebugLogger {
	var l DebugLogger

	l.SetOutput(out)
	l.SetPrefix(prefix)
	l.SetFlags(flag)
	return &l
}

// A impl 'l DebugLogger' verifiabledata/log.Logger
func (l DebugLogger) Error(v ...interface{}) {
	l.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

func (l DebugLogger) Info(v ...interface{}) {
	l.Output(2, fmt.Sprint(v...))
}

func (l DebugLogger) Debug(v ...interface{}) {
	l.Output(2, fmt.Sprint(v...))
}

func (l DebugLogger) Errorf(format string, v ...interface{}) {
	l.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (l DebugLogger) Infof(format string, v ...interface{}) {
	l.Output(2, fmt.Sprintf(format, v...))
}

func (l DebugLogger) Debugf(format string, v ...interface{}) {
	l.Output(2, fmt.Sprintf(format, v...))
}

