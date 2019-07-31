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
	"io/ioutil"
	"log"
)

type silentLogger struct {
	log.Logger
}

func newSilent() *silentLogger {
	l := &silentLogger{
		Logger: log.Logger{},
	}
	l.SetOutput(ioutil.Discard)
	return l
}

// A impl 'l Nologger' qed/log.Logger
func (l *silentLogger) Error(v ...interface{})                 { osExit(1) }
func (l *silentLogger) Info(v ...interface{})                  { return }
func (l *silentLogger) Debug(v ...interface{})                 { return }
func (l *silentLogger) Errorf(format string, v ...interface{}) { osExit(1) }
func (l *silentLogger) Infof(format string, v ...interface{})  { return }
func (l *silentLogger) Debugf(format string, v ...interface{}) { return }

func (l *silentLogger) GetLogger() *log.Logger {
	return &l.Logger
}

func (l *silentLogger) GetLoggerLevel() string {
	return SILENT
}
