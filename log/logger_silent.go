// copyright © 2018 banco bilbao vizcaya argentaria s.a.  all rights reserved.
// use of this source code is governed by an apache 2 license
// that can be found in the license file

package log

import (
	"log"
)

type silentLogger struct {
	log.Logger
}

func newSilent() *silentLogger {
	return &silentLogger{}
}

// A impl 'l Nologger' verifiabledata/log.Logger
func (l silentLogger) Error(v ...interface{})                 { return }
func (l silentLogger) Warn(v ...interface{})                  { return }
func (l silentLogger) Info(v ...interface{})                  { return }
func (l silentLogger) Debug(v ...interface{})                 { return }
func (l silentLogger) Errorf(format string, v ...interface{}) { return }
func (l silentLogger) Warnf(format string, v ...interface{})  { return }
func (l silentLogger) Infof(format string, v ...interface{})  { return }
func (l silentLogger) Debugf(format string, v ...interface{}) { return }
