package log

import (
	"bytes"
	"io"
	"log"

	"github.com/hashicorp/go-hclog"
)

// HclogAdapter implements Hashicorp's hclog.Logger interface using QED's Logger.
// It's a workaround for raft system. Hashicorp's Raft doesn't support other
// logger than hclog. This logger implements only methods used by Raft.
type HclogAdapter struct {
	log Logger
}

func NewHclogAdapter(log Logger) *HclogAdapter {
	return &HclogAdapter{log: log}
}

// Trace implementation
func (l HclogAdapter) Trace(msg string, args ...interface{}) {
	l.log.Tracef("%s: %v", msg, argsToString(args...))
}

// Debug implementation
func (l HclogAdapter) Debug(msg string, args ...interface{}) {
	l.log.Debugf("%s: %v", msg, argsToString(args...))
}

// Info implementation
func (l HclogAdapter) Info(msg string, args ...interface{}) {
	l.log.Infof("%s: %v", msg, argsToString(args...))
}

// Warn implementation
func (l HclogAdapter) Warn(msg string, args ...interface{}) {
	l.log.Warnf("%s: %v", msg, argsToString(args...))
}

// Error implementation
func (l HclogAdapter) Error(msg string, args ...interface{}) {
	l.log.Errorf("%s: %v", msg, argsToString(args...))
}

// IsTrace implementation.
func (l HclogAdapter) IsTrace() bool {
	_, ok := l.log.(*traceLogger)
	return ok
}

// IsDebug implementation.
func (l HclogAdapter) IsDebug() bool {
	_, ok := l.log.(*debugLogger)
	return ok
}

// IsInfo implementation.
func (l HclogAdapter) IsInfo() bool {
	_, ok := l.log.(*infoLogger)
	return ok
}

// IsWarn implementation.
func (l HclogAdapter) IsWarn() bool {
	_, ok := l.log.(*warnLogger)
	return ok
}

// IsError implementation.
func (l HclogAdapter) IsError() bool {
	_, ok := l.log.(*errorLogger)
	return ok
}

// With implementation.
func (l HclogAdapter) With(args ...interface{}) hclog.Logger {
	// no need to implement that as Raft doesn't use this method.
	return l
}

// Named implementation.
func (l HclogAdapter) Named(name string) hclog.Logger {
	return HclogAdapter{log: l.log.Named(name)}
}

// ResetNamed implementation.
func (l HclogAdapter) ResetNamed(name string) hclog.Logger {
	return HclogAdapter{log: l.log.ResetNamed(name)}
}

// SetLevel implementation.
func (l HclogAdapter) SetLevel(level hclog.Level) {
	l.log = l.log.WithLevel(Level(level))
}

// StandardLogger implementation.
func (l HclogAdapter) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	return l.log.StdLogger(&StdLoggerOptions{
		InferLevels: opts.InferLevels,
		ForceLevel:  Level(opts.ForceLevel),
	})
}

// StandardWriter implementation
func (l HclogAdapter) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
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
