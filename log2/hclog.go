package log2

import (
	"bytes"
	"io"
	"log"

	"github.com/hashicorp/go-hclog"
)

// Hclog2Logger implements Hashicorp's hclog.Logger interface using QED's Logger.
// It's a workaround for raft system. Hashicorp's Raft doesn't support other
// logger than hclog. This logger implements only methods used by Raft.
type Hclog2Logger struct {
	log Logger
}

func NewHclog2Logger(log Logger) *Hclog2Logger {
	return &Hclog2Logger{log: log}
}

// Trace implementation
func (l Hclog2Logger) Trace(msg string, args ...interface{}) {
	l.log.Tracef("%s: %v", msg, argsToString(args...))
}

// Debug implementation
func (l Hclog2Logger) Debug(msg string, args ...interface{}) {
	l.log.Debugf("%s: %v", msg, argsToString(args...))
}

// Info implementation
func (l Hclog2Logger) Info(msg string, args ...interface{}) {
	l.log.Infof("%s: %v", msg, argsToString(args...))
}

// Warn implementation
func (l Hclog2Logger) Warn(msg string, args ...interface{}) {
	l.log.Warnf("%s: %v", msg, argsToString(args...))
}

// Error implementation
func (l Hclog2Logger) Error(msg string, args ...interface{}) {
	l.log.Errorf("%s: %v", msg, argsToString(args...))
}

// IsTrace implementation.
func (l Hclog2Logger) IsTrace() bool {
	_, ok := l.log.(*traceLogger)
	return ok
}

// IsDebug implementation.
func (l Hclog2Logger) IsDebug() bool {
	_, ok := l.log.(*debugLogger)
	return ok
}

// IsInfo implementation.
func (l Hclog2Logger) IsInfo() bool {
	_, ok := l.log.(*infoLogger)
	return ok
}

// IsWarn implementation.
func (l Hclog2Logger) IsWarn() bool {
	_, ok := l.log.(*warnLogger)
	return ok
}

// IsError implementation.
func (l Hclog2Logger) IsError() bool {
	_, ok := l.log.(*errorLogger)
	return ok
}

// With implementation.
func (l Hclog2Logger) With(args ...interface{}) hclog.Logger {
	// no need to implement that as Raft doesn't use this method.
	return l
}

// Named implementation.
func (l Hclog2Logger) Named(name string) hclog.Logger {
	return Hclog2Logger{log: l.log.Named(name)}
}

// ResetNamed implementation.
func (l Hclog2Logger) ResetNamed(name string) hclog.Logger {
	return Hclog2Logger{log: l.log.ResetNamed(name)}
}

// SetLevel implementation.
func (l Hclog2Logger) SetLevel(level hclog.Level) {
	// no need to implement that as Raft doesn't use this method.
}

// StandardLogger implementation.
func (l Hclog2Logger) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	return l.log.StdLogger(&StdLoggerOptions{
		InferLevels: opts.InferLevels,
		ForceLevel:  Level(opts.ForceLevel),
	})
}

// StandardWriter implementation
func (l Hclog2Logger) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return l.log.StdWriter(&StdLoggerOptions{
		InferLevels: opts.InferLevels,
		ForceLevel:  Level(opts.ForceLevel),
	})
}

func argsToString(args ...interface{}) string {
	buf := bytes.Buffer{}
	for i := 0; i < len(args); i += 2 {
		buf.WriteString(args[i].(string))
		buf.WriteByte('=')
		buf.WriteString(args[i+1].(string))
	}
	return buf.String()
}
