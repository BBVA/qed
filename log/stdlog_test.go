package log

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStdLogAdapter(t *testing.T) {

	t.Run("use internal level", func(t *testing.T) {

		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:   "test",
			Level:  Info,
			Output: &buf,
		})

		stdLogAdapter := logger.StdLogger(nil)

		stdLogAdapter.Printf("this is a test")

		str := buf.String()
		str = str[strings.IndexByte(str, ' ')+1:]

		require.Equal(t, "[INFO]  test: this is a test\n", str)
	})

	t.Run("use internal but restrictive level", func(t *testing.T) {

		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:   "test",
			Level:  Error,
			Output: &buf,
		})

		stdLogAdapter := logger.StdLogger(nil)

		stdLogAdapter.Printf("this is a test")

		str := buf.String()
		str = str[strings.IndexByte(str, ' ')+1:]

		require.Equal(t, "", str)
	})

	t.Run("infer level", func(t *testing.T) {

		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:   "test",
			Level:  Info,
			Output: &buf,
		})

		stdLogAdapter := logger.StdLogger(&StdLoggerOptions{
			InferLevels: true,
		})

		stdLogAdapter.Printf("[INFO] this is a test")

		str := buf.String()
		str = str[strings.IndexByte(str, ' ')+1:]

		require.Equal(t, "[INFO]  test: this is a test\n", str)
	})

	t.Run("force level to Off", func(t *testing.T) {

		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:   "test",
			Level:  Info,
			Output: &buf,
		})

		stdLogAdapter := logger.StdLogger(&StdLoggerOptions{
			ForceLevel: Off,
		})

		stdLogAdapter.Printf("this is a test")

		str := buf.String()
		str = str[strings.IndexByte(str, ' ')+1:]

		require.Equal(t, "", str)
	})

	t.Run("force level to more restrictive", func(t *testing.T) {

		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:   "test",
			Level:  Info,
			Output: &buf,
		})

		stdLogAdapter := logger.StdLogger(&StdLoggerOptions{
			ForceLevel: Error,
		})

		stdLogAdapter.Printf("this is a test")

		str := buf.String()
		str = str[strings.IndexByte(str, ' ')+1:]

		require.Equal(t, "[ERROR] test: this is a test\n", str)
	})

}
