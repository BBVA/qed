package log

import (
	"bytes"
	"testing"

	"github.com/hashicorp/go-hclog"
)

func BenchmarkLogger(b *testing.B) {
	b.Run("sequential info with error level", func(b *testing.B) {

		var buf bytes.Buffer
		logger := New(&LoggerOptions{
			Name:            "test",
			Level:           Error,
			Output:          &buf,
			TimeFormat:      DefaultTimeFormat,
			IncludeLocation: true,
		})

		for i := 0; i < b.N; i++ {
			logger.Info("this is a test")
		}
	})

	b.Run("sequential info with info level", func(b *testing.B) {

		var buf bytes.Buffer
		logger := New(&LoggerOptions{
			Name:            "test",
			Level:           Info,
			Output:          &buf,
			TimeFormat:      DefaultTimeFormat,
			IncludeLocation: true,
		})

		for i := 0; i < b.N; i++ {
			logger.Info("this is a test")
		}
	})

}

func BenchmarkLoggerParallel(b *testing.B) {

	// parallel info with error level
	var buf bytes.Buffer
	logger := New(&LoggerOptions{
		Name:            "test",
		Level:           Error,
		Output:          &buf,
		TimeFormat:      DefaultTimeFormat,
		IncludeLocation: true,
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("this is a test")
		}

	})

}

func BenchmarkHcLogger(b *testing.B) {
	b.Run("sequential info with error level", func(b *testing.B) {

		var buf bytes.Buffer
		logger := hclog.New(&hclog.LoggerOptions{
			Name:            "test",
			Level:           hclog.Error,
			Output:          &buf,
			TimeFormat:      hclog.TimeFormat,
			IncludeLocation: true,
		})

		for i := 0; i < b.N; i++ {
			logger.Info("this is a test")
		}
	})

	b.Run("sequential info with info level", func(b *testing.B) {

		var buf bytes.Buffer
		logger := hclog.New(&hclog.LoggerOptions{
			Name:            "test",
			Level:           hclog.Info,
			Output:          &buf,
			TimeFormat:      hclog.TimeFormat,
			IncludeLocation: true,
		})

		for i := 0; i < b.N; i++ {
			logger.Info("this is a test")
		}
	})
}

func BenchmarkHcLoggerParallel(b *testing.B) {

	// parallel info with error level
	var buf bytes.Buffer
	logger := hclog.New(&hclog.LoggerOptions{
		Name:            "test",
		Level:           hclog.Error,
		Output:          &buf,
		TimeFormat:      hclog.TimeFormat,
		IncludeLocation: true,
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("this is a test")
		}

	})

}
