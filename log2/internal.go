package log2

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	brackets = map[Level]string{
		Trace: "[TRACE]",
		Debug: "[DEBUG]",
		Info:  "[INFO] ",
		Warn:  "[WARN] ",
		Error: "[ERROR]",
		Fatal: "[FATAL]",
	}
)

type internalLogger struct {
	name       string
	caller     bool
	timeFormat string

	// This is a pointer so that it's shared by any derived loggers, since
	// those derived loggers share the bufio.Writer as well.
	mutex  *sync.Mutex
	writer *writer
}

func (l *internalLogger) Named(name string) Logger {
	subLogger := *l
	if subLogger.name != "" {
		subLogger.name = subLogger.name + "." + name
	} else {
		subLogger.name = name
	}
	return &subLogger
}

func (l *internalLogger) ResetNamed(name string) Logger {
	rl := *l
	rl.name = name
	return &rl
}

func (l *internalLogger) StdLogger(opts *StdLoggerOptions) *log.Logger {
	if opts == nil {
		opts = &StdLoggerOptions{}
	}
	return log.New(l.StdWriter(opts), "", 0)
}

func (l *internalLogger) StdWriter(opts *StdLoggerOptions) io.Writer {
	return &stdLogAdapter{
		log:         l,
		inferLevels: opts.InferLevels,
		forceLevel:  opts.ForceLevel,
	}
}

func (l *internalLogger) log(level Level, msg string) {
	tm := time.Now()

	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.logPlain(tm, level, msg) // TODO we can extend it to logJSON

}

func (l *internalLogger) logf(level Level, format string, args ...interface{}) {
	tm := time.Now()

	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.logPlain(tm, level, fmt.Sprintf(format, args...)) // TODO we can extend it to logJSON

}

func (l *internalLogger) logPlain(tm time.Time, level Level, msg string) {

	// time
	l.writer.WriteString(tm.Format(l.timeFormat))

	// level
	l.writer.WriteByte(' ')
	l.writer.WriteString(levelToBracket(level))

	// caller
	if l.caller {
		if _, file, line, ok := runtime.Caller(4); ok {
			l.writer.WriteByte(' ')
			l.writer.WriteString(trimCallerPath(file))
			l.writer.WriteByte(':')
			l.writer.WriteString(strconv.Itoa(line))
			l.writer.WriteByte(':')
		}
	}

	// name
	l.writer.WriteByte(' ')
	if l.name != "" {
		l.writer.WriteString(l.name)
		l.writer.WriteString(": ")
	}

	// msg
	l.writer.WriteString(msg)

	l.writer.WriteString("\n")
	l.writer.Flush()
}

func trimCallerPath(path string) string {
	// cleanups a path by returning only the last 2 segments of the path.

	// find the last separator
	var idx int
	if idx = strings.LastIndexByte(path, '/'); idx == -1 {
		return path
	}

	// find the penultimate separator
	if idx = strings.LastIndexByte(path[:idx], '/'); idx == -1 {
		return path
	}

	return path[idx+1:]

}

func levelToBracket(level Level) string {
	s, ok := brackets[level]
	if !ok {
		s = "[?????]"
	}
	return s
}

func (l *internalLogger) Trace(msg string) {
	l.log(Trace, msg)
}

func (l *internalLogger) Tracef(format string, args ...interface{}) {
	l.logf(Trace, format, args...)
}

func (l *internalLogger) Debug(msg string) {
	l.log(Debug, msg)
}

func (l *internalLogger) Debugf(format string, args ...interface{}) {
	l.logf(Debug, format, args...)
}

func (l *internalLogger) Info(msg string) {
	l.log(Info, msg)
}

func (l *internalLogger) Infof(format string, args ...interface{}) {
	l.logf(Info, format, args...)
}

func (l *internalLogger) Warn(msg string) {
	l.log(Warn, msg)
}

func (l *internalLogger) Warnf(format string, args ...interface{}) {
	l.logf(Warn, format, args...)
}

func (l *internalLogger) Error(msg string) {
	l.log(Error, msg)
}

func (l *internalLogger) Errorf(format string, args ...interface{}) {
	l.logf(Error, format, args...)
}

func (l *internalLogger) Fatal(msg string) {
	l.log(Error, msg)
	os.Exit(1)
}

func (l *internalLogger) Fatalf(format string, args ...interface{}) {
	l.logf(Error, format, args...)
	os.Exit(1)
}

func (l *internalLogger) Panic(msg string) {
	l.log(Error, msg)
	panic(msg)
}

func (l *internalLogger) Panicf(format string, args ...interface{}) {
	l.logf(Error, format, args...)
	panic(fmt.Sprintf(format, args...))
}
