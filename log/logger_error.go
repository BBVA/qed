// copyright © 2018 banco bilbao vizcaya argentaria s.a.  all rights reserved.
// use of this source code is governed by an apache 2 license
// that can be found in the license file

package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

type errorLogger struct {
	log.Logger
}

func newError(out io.Writer, prefix string, flag int) *errorLogger {
	var l errorLogger

	l.SetOutput(out)
	l.SetPrefix(prefix)
	l.SetFlags(flag)
	return &l
}

// A impl 'l errorLogger' verifiabledata/log.Logger
func (l errorLogger) Error(v ...interface{}) {
	l.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

func (l errorLogger) Errorf(format string, v ...interface{}) {
	l.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (l errorLogger) Info(v ...interface{})                  { return }
func (l errorLogger) Debug(v ...interface{})                 { return }
func (l errorLogger) Infof(format string, v ...interface{})  { return }
func (l errorLogger) Debugf(format string, v ...interface{}) { return }
