package log2

import (
	"io"
	"os"
	"sync"
)

// DefaultTimeFormat to use for logging. This is a version of RFC3339 that contains
// contains millisecond precision.
const DefaultTimeFormat = "2006-01-02T15:04:05.000Z07:00"

var (
	defLock   sync.Once
	defLogger Logger

	//DefaultOutput is used as the default log output.
	DefaultOutput io.Writer = os.Stderr

	// DefaultLevel is used as the default log level.
	DefaultLevel = Off
)

// Default is used to create a default logger.
// Once the logger is created, these options are ignored,
// so set them as soon as the process starts.
func Default() Logger {
	defLock.Do(func() {
		if defLogger == nil {
			defLogger = New(&LoggerOptions{
				Level:  DefaultLevel,
				Output: DefaultOutput,
			})
		}
	})
	return defLogger
}

func L() Logger {
	return Default()
}

func SetDefault(log Logger) Logger {
	prev := defLogger
	defLogger = log
	return prev
}
