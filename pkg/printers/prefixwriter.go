package printers

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type flusher interface {
	Flush() error
}

// NewPrefixWriter creates a new PrefixWriter.
func NewPrefixWriter(out io.Writer) PrefixWriter {
	tabWriter := tabwriter.NewWriter(out, 0, 8, 2, ' ', 0)
	return &prefixWriter{out: tabWriter}
}

// PrefixWriter can write text at various indentation levels.
type PrefixWriter interface {
	// Write writes text with the specified indentation level.
	Write(level int, format string, a ...interface{})
	// WriteLine writes an entire line with no indentation level.
	WriteLine(a ...interface{})
	// Write columns with an initial indentation
	WriteCols(level int, cols ...string)
	// Write columns with an initial indentation and a newline at the end
	WriteColsLn(level int, cols ...string)
	// Flush forces indentation to be reset.
	Flush() error
}

// prefixWriter implements PrefixWriter
type prefixWriter struct {
	out io.Writer
}

var _ PrefixWriter = &prefixWriter{}

// Each level has 2 spaces for PrefixWriter
const (
	LEVEL_0 = iota
	LEVEL_1
	LEVEL_2
	LEVEL_3
)

func (pw *prefixWriter) Write(level int, format string, a ...interface{}) {
	levelSpace := "  "
	prefix := ""
	for i := 0; i < level; i++ {
		prefix += levelSpace
	}
	fmt.Fprintf(pw.out, prefix+format, a...)
}

func (pw *prefixWriter) WriteCols(level int, cols ...string) {
	ss := make([]string, len(cols))
	for i, _ := range cols {
		ss[i] = "%s"
	}
	format := strings.Join(ss, "\t")
	s := make([]interface{}, len(cols))
	for i, v := range cols {
		s[i] = v
	}

	pw.Write(level, format, s...)
}

func (pw *prefixWriter) WriteColsLn(level int, cols ...string) {
	pw.WriteCols(level, cols...)
	pw.WriteLine()
}

func (pw *prefixWriter) WriteLine(a ...interface{}) {
	fmt.Fprintln(pw.out, a...)
}

func (pw *prefixWriter) Flush() error {
	if f, ok := pw.out.(flusher); ok {
		return f.Flush()
	}
	return fmt.Errorf("outpustream %v doesn't support Flush", pw.out)
}
