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

type infoLogger struct {
	log.Logger
}

func newInfo(out io.Writer, prefix string, flag int) *infoLogger {
	l := infoLogger{
		Logger: log.Logger{},
	}

	l.SetOutput(out)
	l.SetPrefix(prefix)
	l.SetFlags(flag)
	return &l
}

// A impl 'l infoLogger' qed/log.Logger
func (l *infoLogger) Error(v ...interface{}) {
	l.Output(caller, fmt.Sprint(v...))
	osExit(1)
}

func (l *infoLogger) Info(v ...interface{}) {
	l.Output(caller, fmt.Sprint(v...))
}

func (l *infoLogger) Errorf(format string, v ...interface{}) {
	l.Output(caller, fmt.Sprintf(format, v...))
	osExit(1)
}

func (l *infoLogger) Infof(format string, v ...interface{}) {
	l.Output(caller, fmt.Sprintf(format, v...))
}

func (l *infoLogger) Debug(v ...interface{})                 { return }
func (l *infoLogger) Debugf(format string, v ...interface{}) { return }

func (l *infoLogger) GetLogger() *log.Logger {
	return &l.Logger
}

func (l *infoLogger) GetLoggerLevel() string {
	return INFO
}
