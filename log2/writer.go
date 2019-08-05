package log2

import (
	"bytes"
	"io"
)

type writer struct {
	buf bytes.Buffer
	out io.Writer
}

func newWriter(w io.Writer) *writer {
	return &writer{out: w}
}

func (w *writer) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func (w *writer) WriteByte(c byte) error {
	return w.buf.WriteByte(c)
}

func (w *writer) WriteString(s string) (int, error) {
	return w.buf.WriteString(s)
}

func (w *writer) Flush() (err error) {
	_, err = w.out.Write(w.buf.Bytes())
	w.buf.Reset()
	return err
}
