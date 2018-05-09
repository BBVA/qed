package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

type debugLogger struct {
	log.Logger
}

func newDebug(out io.Writer, prefix string, flag int) *debugLogger {
	var l debugLogger

	l.SetOutput(out)
	l.SetPrefix(prefix)
	l.SetFlags(flag)
	return &l
}

// A impl 'l debugLogger' qed/log.Logger
func (l debugLogger) Error(v ...interface{}) {
	l.Output(caller, fmt.Sprint(v...))
	os.Exit(1)
}

func (l debugLogger) Info(v ...interface{}) {
	l.Output(caller, fmt.Sprint(v...))
}

func (l debugLogger) Debug(v ...interface{}) {
	l.Output(caller, fmt.Sprint(v...))
}

func (l debugLogger) Errorf(format string, v ...interface{}) {
	l.Output(caller, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (l debugLogger) Infof(format string, v ...interface{}) {
	l.Output(caller, fmt.Sprintf(format, v...))
}

func (l debugLogger) Debugf(format string, v ...interface{}) {
	l.Output(caller, fmt.Sprintf(format, v...))
}
