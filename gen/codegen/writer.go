package codegen

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Writer struct {
	buf    *bytes.Buffer
	writer io.Writer
	indent int
}

func (t *Writer) Init() {
	t.buf = &bytes.Buffer{}
	t.writer = t.buf
}

func (t *Writer) String() string {
	return t.buf.String()
}

func (t *Writer) N(format string, i ...any) (err error) {
	new := fmt.Sprintf(format, i...)
	_, err = t.writer.Write([]byte(new))
	if err != nil {
		return err
	}
	return nil
}

func (t *Writer) Indent() string {
	return strings.Repeat("\t", t.indent)
}

func (t *Writer) W(format string, i ...any) (err error) {
	t.writer.Write([]byte(t.Indent()))

	new := fmt.Sprintf(format, i...)
	_, err = t.writer.Write([]byte(new))
	if err != nil {
		return err
	}
	return nil
}

func (t *Writer) IndentIn() {
	t.indent++
}

func (t *Writer) IndentOut() {
	t.indent--
}
