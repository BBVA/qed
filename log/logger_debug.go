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

type debugLogger struct {
	log.Logger
}

func newDebug(out io.Writer, prefix string, flag int) *debugLogger {
	l := debugLogger{
		Logger: log.Logger{},
	}

	l.SetOutput(out)
	l.SetPrefix(prefix)
	l.SetFlags(flag)
	return &l
}

// A impl 'l debugLogger' qed/log.Logger
func (l *debugLogger) Error(v ...interface{}) {
	l.Output(caller, fmt.Sprint(v...))
	osExit(1)
}

func (l *debugLogger) Info(v ...interface{}) {
	l.Output(caller, fmt.Sprint(v...))
}

func (l *debugLogger) Debug(v ...interface{}) {
	l.Output(caller, fmt.Sprint(v...))
}

func (l *debugLogger) Errorf(format string, v ...interface{}) {
	l.Output(caller, fmt.Sprintf(format, v...))
	osExit(1)
}

func (l *debugLogger) Infof(format string, v ...interface{}) {
	l.Output(caller, fmt.Sprintf(format, v...))
}

func (l *debugLogger) Debugf(format string, v ...interface{}) {
	l.Output(caller, fmt.Sprintf(format, v...))
}

func (l *debugLogger) GetLogger() *log.Logger {
	return &l.Logger
}

func (l *debugLogger) GetLoggerLevel() string {
	return DEBUG
}
