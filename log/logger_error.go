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

package log

import (
	"fmt"
	"io"
	"log"
)

type errorLogger struct {
	log.Logger
}

func newError(out io.Writer, prefix string, flag int) *errorLogger {
	l := errorLogger{
		Logger: log.Logger{},
	}

	l.SetOutput(out)
	l.SetPrefix(prefix)
	l.SetFlags(flag)
	return &l
}

// A impl 'l errorLogger' qed/log.Logger
func (l *errorLogger) Error(v ...interface{}) {
	l.Output(caller, fmt.Sprint(v...))
	osExit(1)
}

func (l *errorLogger) Errorf(format string, v ...interface{}) {
	l.Output(caller, fmt.Sprintf(format, v...))
	osExit(1)
}

func (l *errorLogger) Info(v ...interface{})                  { return }
func (l *errorLogger) Debug(v ...interface{})                 { return }
func (l *errorLogger) Infof(format string, v ...interface{})  { return }
func (l *errorLogger) Debugf(format string, v ...interface{}) { return }

func (l *errorLogger) GetLogger() *log.Logger {
	return &l.Logger
}

func (l *errorLogger) GetLoggerLevel() string {
	return ERROR
}
