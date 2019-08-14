package log2

import (
	"bytes"
	"strings"
)

type stdLogAdapter struct {
	log         Logger
	inferLevels bool
	forceLevel  Level
}

func (s *stdLogAdapter) Write(data []byte) (int, error) {

	str := string(bytes.TrimRight(data, " \t\n"))

	if s.forceLevel != NotSet {
		// use pickLevel ot strip log levels included in the line since
		// we are forcing the level
		_, str := s.pickLevel(str)

		switch s.forceLevel {
		case Off:
		case Trace:
			s.log.Trace(str)
		case Debug:
			s.log.Debug(str)
		case Info:
			s.log.Info(str)
		case Warn:
			s.log.Warn(str)
		case Error:
			s.log.Error(str)
		case Fatal:
			s.log.Fatal(str)
		default:
			s.log.Info(str)
		}
	} else if s.inferLevels {
		level, str := s.pickLevel(str)
		switch level {
		case Trace:
			s.log.Trace(str)
		case Debug:
			s.log.Debug(str)
		case Info:
			s.log.Info(str)
		case Warn:
			s.log.Warn(str)
		case Error:
			s.log.Error(str)
		default:
			s.log.Info(str)
		}
	} else {
		s.log.Info(str)
	}

	return len(data), nil
}

func (s *stdLogAdapter) pickLevel(str string) (Level, string) {
	// Detect, based on conventions, what log level this is.
	// NOTE: this is useful to capture Hashicorp's Memberlist traces.
	switch {
	case strings.HasPrefix(str, "[TRACE]"):
		return Trace, strings.TrimSpace(str[7:])
	case strings.HasPrefix(str, "[DEBUG]"):
		return Debug, strings.TrimSpace(str[7:])
	case strings.HasPrefix(str, "[INFO]"):
		return Info, strings.TrimSpace(str[6:])
	case strings.HasPrefix(str, "[WARN]"):
		return Warn, strings.TrimSpace(str[7:])
	case strings.HasPrefix(str, "[ERROR]"):
		return Error, strings.TrimSpace(str[7:])
	case strings.HasPrefix(str, "[ERR]"):
		return Error, strings.TrimSpace(str[5:])
	case strings.HasPrefix(str, "[FATAL]"):
		return Fatal, strings.TrimSpace(str[7:])
	default:
		return Info, str
	}
}
