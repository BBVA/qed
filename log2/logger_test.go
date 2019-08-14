package log2

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {

	t.Run("default", func(t *testing.T) {

		var buf bytes.Buffer
		SetDefault(New(&LoggerOptions{
			Output: &buf,
			Level:  Info,
		}))

		L().Info("this is a test")

		str := buf.String()
		str = str[strings.IndexByte(str, ' ')+1:]

		require.Equal(t, "[INFO]  this is a test\n", str)
	})

	t.Run("default with error level", func(t *testing.T) {

		var buf bytes.Buffer
		SetDefault(New(&LoggerOptions{
			Output: &buf,
			Level:  Error,
		}))

		L().Info("this is a test")

		require.Equal(t, "", buf.String())

		L().Error("this is a test")

		str := buf.String()
		str = str[strings.IndexByte(str, ' ')+1:]

		require.Equal(t, "[ERROR] this is a test\n", str)
	})

	t.Run("named from default", func(t *testing.T) {

		var buf bytes.Buffer
		SetDefault(New(&LoggerOptions{
			Name:   "default",
			Output: &buf,
			Level:  Info,
		}))

		L().Named("test").Info("this is a test")

		str := buf.String()
		str = str[strings.IndexByte(str, ' ')+1:]

		require.Equal(t, "[INFO]  default.test: this is a test\n", str)

	})

	t.Run("uses default output if none is given", func(t *testing.T) {

		var buf bytes.Buffer
		DefaultOutput = &buf

		logger := New(&LoggerOptions{
			Name:  "test",
			Level: Info,
		})

		logger.Info("this is a test")

		str := buf.String()
		str = str[strings.IndexByte(str, ' ')+1:]

		require.Equal(t, "[INFO]  test: this is a test\n", str)

	})

	t.Run("includes the caller location", func(t *testing.T) {

		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:            "test",
			Output:          &buf,
			Level:           Info,
			IncludeLocation: true,
		})

		logger.Info("this is a test")

		str := buf.String()
		str = str[strings.IndexByte(str, ' ')+1:]

		// This test will break if you move this around, it's line dependent
		require.Equal(t, "[INFO]  log2/logger_test.go:97: test: this is a test\n", str)

	})
}
