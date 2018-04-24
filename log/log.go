package log

import (
	"log"
)

type Logger interface {
	Error(v ...interface{})
	Info(v ...interface{})
	Debugf(v ...interface{})
}

type ErrorLogger struct {
	log.Logger
}

func (l ErrorLogger) Error(v ...interface{}) {
	l.Printf(v[0].(string), v[1:])
}

func (l ErrorLogger) Info(v ...interface{}) {
	return
}

func (l ErrorLogger) Debugf(v ...interface{}) {
	return
}

type InfoLogger struct {
	log.Logger
}

func (l InfoLogger) Error(v ...interface{}) {
	l.Printf(v[0].(string), v[1:])
}

func (l InfoLogger) Info(v ...interface{}) {
	l.Printf(v[0].(string), v[1:])
}

func (l InfoLogger) Debugf(v ...interface{}) {
	return
}

type DebugLogger struct {
	log.Logger
}

func (l DebugLogger) Info(v ...interface{}) {
	l.Printf(v[0].(string), v[1:])
}

func (l DebugLogger) Error(v ...interface{}) {
	l.Printf(v[0].(string), v[1:])
}

func (l DebugLogger) Debugf(v ...interface{}) {
	l.Printf(v[0].(string), v[1:])
}
