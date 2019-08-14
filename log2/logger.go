package log2

import (
	"io"
	"log"
	"strings"
	"sync"
)

// Level represents the logging level.
type Level uint32

const (
	// NotSet level is used to indicate that no level has been set
	// and allow for a default to be used
	NotSet Level = iota

	// Off is intended to avoid tracing any action.
	Off

	// Fatal
	Fatal

	// Error
	Error

	// Warn
	Warn

	// Info
	Info

	// Debug
	Debug

	// Trace
	Trace
)

func (l Level) String() string {
	switch l {
	case NotSet:
		return "unknown"
	case Off:
		return "off"
	case Fatal:
		return "fatal"
	case Error:
		return "error"
	case Warn:
		return "warn"
	case Info:
		return "info"
	case Debug:
		return "debug"
	case Trace:
		return "trace"
	default:
		return "unknown"
	}
}

// LevelFromString returns a Level type for the named log level, or
// "NotSet" if the level passed as argument is invalid.
func LevelFromString(level string) Level {
	level = strings.ToLower(strings.TrimSpace(level))
	switch level {
	case "off":
		return Off
	case "fatal":
		return Fatal
	case "error":
		return Error
	case "warn":
		return Warn
	case "info":
		return Info
	case "debug":
		return Debug
	case "trace":
		return Trace
	default:
		return NotSet
	}
}

type Logger interface {
	Trace(msg string)
	Tracef(format string, args ...interface{})
	Debug(msg string)
	Debugf(format string, args ...interface{})
	Info(msg string)
	Infof(format string, args ...interface{})
	Warn(msg string)
	Warnf(format string, args ...interface{})
	Error(msg string)
	Errorf(format string, args ...interface{})
	Fatal(msg string)
	Fatalf(format string, args ...interface{})
	Panic(msg string)
	Panicf(format string, args ...interface{})

	// Create a logger that will prepend the given name on front of all
	// messages. If the logger has a previously set name, the new value
	// will be the appended to it.
	Named(name string) Logger

	// Create a logger that will prepend the given name on front of all
	// messages. It overrides any previously set name.
	ResetNamed(name string) Logger

	WithLevel(level Level) Logger

	// StdLogger returns a logger implementation that conforms to the
	// stdlib log.Logger interface. This allows packages that expect
	// to be using the standard library log to actually use this logger.
	StdLogger(opts *StdLoggerOptions) *log.Logger

	// StdWriter returns a io.Writer implementation that conforms to
	// io.Writer, which can be passed into log.SetOutput().
	StdWriter(opts *StdLoggerOptions) io.Writer
}

// LoggerOptions can be used to configure a new logger.
type LoggerOptions struct {
	// Name of the subsystem to prefix logs with.
	Name string

	// Level is the threshold for the logger. Any log trace less
	// sever is supressed.
	Level Level

	// Output is the writer implementation where to write logs to.
	// If nil, defaults to os.Stderr.
	Output io.Writer

	// TimeFormat is the time format to use instead of the default one.
	TimeFormat string

	// IncludeLocation includes file and line information in each log line.
	IncludeLocation bool

	// Mutex is an optional mutex pointer in case Output is shared.
	Mutex *sync.Mutex
}

// StdLoggerOptions can be used to configure a new standard logger.
type StdLoggerOptions struct {
	// Indicate that some minimal parsing should be done on strings to try
	// and detect their level and re-emit them.
	// This supports the strings like [FATAL], [ERROR], [TRACE], [WARN], [INFO],
	// [DEBUG] and strip it off before reapplying it.
	InferLevels bool

	// ForceLevel is used to force all output from the standard logger to be at
	// the specified level. Similar to InferLevels, this will strip any level
	// prefix contained in the logged string before applying the forced level.
	// If set, this override InferLevels.
	ForceLevel Level
}

func New(opts *LoggerOptions) Logger {
	if opts == nil {
		opts = &LoggerOptions{}
	}

	output := opts.Output
	if output == nil {
		output = DefaultOutput
	}

	level := opts.Level
	if level == NotSet {
		level = DefaultLevel
	}

	mutex := opts.Mutex
	if mutex == nil {
		mutex = new(sync.Mutex)
	}

	timeFormat := opts.TimeFormat
	if timeFormat == "" {
		timeFormat = DefaultTimeFormat
	}

	intLogger := internalLogger{
		name:       opts.Name,
		caller:     opts.IncludeLocation,
		timeFormat: opts.TimeFormat,
		level:      opts.Level,
		mutex:      mutex,
		writer:     newWriter(output),
	}

	var l Logger
	switch level {
	case Off:
		l = &silentLogger{intLogger}
	case Fatal:
		l = &fatalLogger{intLogger}
	case Error:
		l = &errorLogger{intLogger}
	case Warn:
		l = &warnLogger{intLogger}
	case Info:
		l = &infoLogger{intLogger}
	case Debug:
		l = &debugLogger{intLogger}
	case Trace:
		l = &traceLogger{intLogger}
	default:
		l = &infoLogger{intLogger}
	}

	return l
}
