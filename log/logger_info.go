package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

type infoLogger struct {
	log.Logger
}

func newInfo(out io.Writer, prefix string, flag int) *infoLogger {
	var l infoLogger

	l.SetOutput(out)
	l.SetPrefix(prefix)
	l.SetFlags(flag)
	return &l
}

// A impl 'l infoLogger' qed/log.Logger
func (l infoLogger) Error(v ...interface{}) {
	l.Output(caller, fmt.Sprint(v...))
	os.Exit(1)
}

func (l infoLogger) Info(v ...interface{}) {
	l.Output(caller, fmt.Sprint(v...))
}

func (l infoLogger) Errorf(format string, v ...interface{}) {
	l.Output(caller, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (l infoLogger) Infof(format string, v ...interface{}) {
	l.Output(caller, fmt.Sprintf(format, v...))
}

func (l infoLogger) Debug(v ...interface{})                 { return }
func (l infoLogger) Debugf(format string, v ...interface{}) { return }
