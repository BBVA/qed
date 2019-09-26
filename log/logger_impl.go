package log

import (
	"fmt"
	"os"
)

type silentLogger struct {
	internalLogger
}

func (l *silentLogger) Trace(msg string)                          {}
func (l *silentLogger) Tracef(format string, args ...interface{}) {}
func (l *silentLogger) Debug(msg string)                          {}
func (l *silentLogger) Debugf(format string, args ...interface{}) {}
func (l *silentLogger) Info(msg string)                           {}
func (l *silentLogger) Infof(format string, args ...interface{})  {}
func (l *silentLogger) Warn(msg string)                           {}
func (l *silentLogger) Warnf(format string, args ...interface{})  {}
func (l *silentLogger) Error(msg string)                          {}
func (l *silentLogger) Errorf(format string, args ...interface{}) {}
func (l *silentLogger) Fatal(msg string) {
	os.Exit(1)
}
func (l *silentLogger) Fatalf(format string, args ...interface{}) {
	os.Exit(1)
}
func (l *silentLogger) Panic(msg string) {
	panic(msg)
}
func (l *silentLogger) Panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

type fatalLogger struct {
	internalLogger
}

func (l *fatalLogger) Trace(msg string)                          {}
func (l *fatalLogger) Tracef(format string, args ...interface{}) {}
func (l *fatalLogger) Debug(msg string)                          {}
func (l *fatalLogger) Debugf(format string, args ...interface{}) {}
func (l *fatalLogger) Info(msg string)                           {}
func (l *fatalLogger) Infof(format string, args ...interface{})  {}
func (l *fatalLogger) Warn(msg string)                           {}
func (l *fatalLogger) Warnf(format string, args ...interface{})  {}
func (l *fatalLogger) Error(msg string)                          {}
func (l *fatalLogger) Errorf(format string, args ...interface{}) {}
func (l *fatalLogger) Fatal(msg string) {
	l.internalLogger.Error(msg)
	os.Exit(1)
}
func (l *fatalLogger) Fatalf(format string, args ...interface{}) {
	l.internalLogger.Errorf(format, args...)
	os.Exit(1)
}
func (l *fatalLogger) Panic(msg string) {
	l.internalLogger.Error(msg)
	panic(msg)
}
func (l *fatalLogger) Panicf(format string, args ...interface{}) {
	l.internalLogger.Errorf(format, args...)
	panic(fmt.Sprintf(format, args...))
}

type errorLogger struct {
	internalLogger
}

func (l *errorLogger) Trace(msg string)                          {}
func (l *errorLogger) Tracef(format string, args ...interface{}) {}
func (l *errorLogger) Debug(msg string)                          {}
func (l *errorLogger) Debugf(format string, args ...interface{}) {}
func (l *errorLogger) Info(msg string)                           {}
func (l *errorLogger) Infof(format string, args ...interface{})  {}
func (l *errorLogger) InfoMsg(msg string)                        {}
func (l *errorLogger) Warn(msg string)                           {}
func (l *errorLogger) Warnf(format string, args ...interface{})  {}
func (l *errorLogger) Error(msg string) {
	l.internalLogger.Error(msg)
}
func (l *errorLogger) Errorf(format string, args ...interface{}) {
	l.internalLogger.Errorf(format, args...)
}
func (l *errorLogger) Fatal(msg string) {
	l.internalLogger.Error(msg)
	os.Exit(1)
}
func (l *errorLogger) Fatalf(format string, args ...interface{}) {
	l.internalLogger.Errorf(format, args...)
	os.Exit(1)
}
func (l *errorLogger) Panic(msg string) {
	l.internalLogger.Error(msg)
	panic(msg)
}
func (l *errorLogger) Panicf(format string, args ...interface{}) {
	l.internalLogger.Errorf(format, args...)
	panic(fmt.Sprintf(format, args...))
}

type warnLogger struct {
	internalLogger
}

func (l *warnLogger) Trace(msg string)                          {}
func (l *warnLogger) Tracef(format string, args ...interface{}) {}
func (l *warnLogger) Debug(msg string)                          {}
func (l *warnLogger) Debugf(format string, args ...interface{}) {}
func (l *warnLogger) Info(msg string)                           {}
func (l *warnLogger) Infof(format string, args ...interface{})  {}
func (l *warnLogger) Warn(msg string) {
	l.internalLogger.Warn(msg)
}
func (l *warnLogger) Warnf(format string, args ...interface{}) {
	l.internalLogger.Warnf(format, args...)
}
func (l *warnLogger) Error(msg string) {
	l.internalLogger.Error(msg)
}
func (l *warnLogger) Errorf(format string, args ...interface{}) {
	l.internalLogger.Errorf(format, args...)
}
func (l *warnLogger) Fatal(msg string) {
	l.internalLogger.Error(msg)
	os.Exit(1)
}
func (l *warnLogger) Fatalf(format string, args ...interface{}) {
	l.internalLogger.Errorf(format, args...)
	os.Exit(1)
}
func (l *warnLogger) Panic(msg string) {
	l.internalLogger.Error(msg)
	panic(msg)
}
func (l *warnLogger) Panicf(format string, args ...interface{}) {
	l.internalLogger.Errorf(format, args...)
	panic(fmt.Sprintf(format, args...))
}

type infoLogger struct {
	internalLogger
}

func (l *infoLogger) Trace(msg string)                          {}
func (l *infoLogger) Tracef(format string, args ...interface{}) {}
func (l *infoLogger) Debug(msg string)                          {}
func (l *infoLogger) Debugf(format string, args ...interface{}) {}
func (l *infoLogger) Info(msg string) {
	l.internalLogger.Info(msg)
}
func (l *infoLogger) Infof(format string, args ...interface{}) {
	l.internalLogger.Infof(format, args...)
}
func (l *infoLogger) Warn(msg string) {
	l.internalLogger.Warn(msg)
}
func (l *infoLogger) Warnf(format string, args ...interface{}) {
	l.internalLogger.Warnf(format, args...)
}
func (l *infoLogger) Error(msg string) {
	l.internalLogger.Error(msg)
}
func (l *infoLogger) Errorf(format string, args ...interface{}) {
	l.internalLogger.Errorf(format, args...)
}
func (l *infoLogger) Fatal(msg string) {
	l.internalLogger.Error(msg)
	os.Exit(1)
}
func (l *infoLogger) Fatalf(format string, args ...interface{}) {
	l.internalLogger.Errorf(format, args...)
	os.Exit(1)
}
func (l *infoLogger) Panic(msg string) {
	l.internalLogger.Error(msg)
	panic(msg)
}
func (l *infoLogger) Panicf(format string, args ...interface{}) {
	l.internalLogger.Errorf(format, args...)
	panic(fmt.Sprintf(format, args...))
}

type debugLogger struct {
	internalLogger
}

func (l *debugLogger) Trace(msg string)                          {}
func (l *debugLogger) Tracef(format string, args ...interface{}) {}
func (l *debugLogger) Debug(msg string) {
	l.internalLogger.Debug(msg)
}
func (l *debugLogger) Debugf(format string, args ...interface{}) {
	l.internalLogger.Debugf(format, args...)
}
func (l *debugLogger) Info(msg string) {
	l.internalLogger.Info(msg)
}
func (l *debugLogger) Infof(format string, args ...interface{}) {
	l.internalLogger.Infof(format, args...)
}
func (l *debugLogger) Warn(msg string) {
	l.internalLogger.Warn(msg)
}
func (l *debugLogger) Warnf(format string, args ...interface{}) {
	l.internalLogger.Warnf(format, args...)
}
func (l *debugLogger) Error(msg string) {
	l.internalLogger.Error(msg)
}
func (l *debugLogger) Errorf(format string, args ...interface{}) {
	l.internalLogger.Errorf(format, args...)
}
func (l *debugLogger) Fatal(msg string) {
	l.internalLogger.Error(msg)
	os.Exit(1)
}
func (l *debugLogger) Fatalf(format string, args ...interface{}) {
	l.internalLogger.Errorf(format, args...)
	os.Exit(1)
}
func (l *debugLogger) Panic(msg string) {
	l.internalLogger.Error(msg)
	panic(msg)
}
func (l *debugLogger) Panicf(format string, args ...interface{}) {
	l.internalLogger.Errorf(format, args...)
	panic(fmt.Sprintf(format, args...))
}

type traceLogger struct {
	internalLogger
}

func (l *traceLogger) Trace(msg string) {
	l.internalLogger.Trace(msg)
}
func (l *traceLogger) Tracef(format string, args ...interface{}) {
	l.internalLogger.Tracef(format, args...)
}
func (l *traceLogger) Debug(msg string) {
	l.internalLogger.Debug(msg)
}
func (l *traceLogger) Debugf(format string, args ...interface{}) {
	l.internalLogger.Debugf(format, args...)
}
func (l *traceLogger) Info(msg string) {
	l.internalLogger.Info(msg)
}
func (l *traceLogger) Infof(format string, args ...interface{}) {
	l.internalLogger.Infof(format, args...)
}
func (l *traceLogger) Warn(msg string) {
	l.internalLogger.Warn(msg)
}
func (l *traceLogger) Warnf(format string, args ...interface{}) {
	l.internalLogger.Warnf(format, args...)
}
func (l *traceLogger) Error(msg string) {
	l.internalLogger.Error(msg)
}
func (l *traceLogger) Errorf(format string, args ...interface{}) {
	l.internalLogger.Errorf(format, args...)
}
func (l *traceLogger) Fatal(msg string) {
	l.internalLogger.Error(msg)
	os.Exit(1)
}
func (l *traceLogger) Fatalf(format string, args ...interface{}) {
	l.internalLogger.Errorf(format, args...)
	os.Exit(1)
}
func (l *traceLogger) Panic(msg string) {
	l.internalLogger.Error(msg)
	panic(msg)
}
func (l *traceLogger) Panicf(format string, args ...interface{}) {
	l.internalLogger.Errorf(format, args...)
	panic(fmt.Sprintf(format, args...))
}
