package grill

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
)

// Test is a single grill test. It is comprised of documentation, commands, and
// expected test results.
type Test struct {
	doc        [][]byte
	command    [][]byte
	expResults [][]byte
	obsResults [][]byte
	changes    []*Change
}

func (t *Test) Failed() bool {
	return len(t.changes) > 0
}

func (t *Test) Skipped() bool {
	return len(t.command) == 0
}

// Status returns a string that represents the suite's overall status.
//
// It is normally used by runner to report the run progress.
func (suite TestSuite) Status() string {
	if suite.Failed() {
		return "failed"
	} else if suite.Skipped() {
		return "skipped"
	} else {
		return "passed"
	}
}

// StatusGlyph returns a one character string that
// represents the suite's overall status.
//
// It is normally used by runner to report the run progress.
func (suite TestSuite) StatusGlyph() string {
	if suite.Failed() {
		return "!"
	} else if suite.Skipped() {
		return "s"
	} else {
		return "."
	}
}

// A TestSuite represents a single grill test file.
type TestSuite struct {
	Name  string
	Dir   string
	Tests []Test
}

// Failed returns true if any test in the suite failed.
func (suite TestSuite) Failed() bool {
	for _, t := range suite.Tests {
		if t.Failed() {
			return true
		}
	}
	return false
}

// Skipped returns true if all of the tests in the suite were skipped.
func (suite TestSuite) Skipped() bool {
	for _, t := range suite.Tests {
		if !t.Skipped() {
			return false
		}
	}
	return true
}

// WriteErr writes test.t.err to the directory that test.t is in.
func (suite TestSuite) WriteErr() error {
	tErr := suite.Name + ".err"
	f, err := os.Create(suite.Name + ".err")
	if err != nil {
		return fmt.Errorf("couldn't write %s: %s", tErr, err)
	}
	for _, t := range suite.Tests {
		for _, d := range t.doc {
			if _, err := fmt.Fprintln(f, string(d)); err != nil {
				return fmt.Errorf("couldn't write %s: %s", tErr, err)
			}
		}
		for i, line := range t.command {
			var format string
			if i == 0 {
				format = "  $ %s\n"
			} else {
				format = "  > %s\n"
			}
			if _, err := fmt.Fprintf(f, format, line); err != nil {
				return fmt.Errorf("couldn't write %s: %s", tErr, err)
			}
		}
		for _, line := range t.obsResults {
			if _, err := fmt.Fprintf(f, "  %s\n", escape(line)); err != nil {
				return fmt.Errorf("couldn't write %s: %s", tErr, err)
			}
		}
	}
	return nil
}

// RemoveErr removes the matching .err file, if it exists.
func (suite TestSuite) RemoveErr() error {
	if err := os.Remove(suite.Name + ".err"); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// WriteReport writes out a report on the overall grill run.
//
// Setting quiet to true will hide the suite diffs
// and write out just the status summary.
func WriteReport(w io.Writer, suites []*TestSuite, ctxLen int, quiet bool) error {
	tests, failed, skipped := 0, 0, 0

	for _, s := range suites {
		if s.Failed() {
			failed++
			if !quiet {
				if err := s.WriteDiff(w, ctxLen); err != nil {
					return fmt.Errorf("couldn't write %q: %s", s.Name+".err", err)
				}
			}
		} else if s.Skipped() {
			skipped++
		}
		tests++
	}

	plural := "s"
	if tests == 1 {
		plural = ""
	}

	_, err := fmt.Fprintf(w, "# Ran %d test%s, %d skipped, %d failed.\n", tests, plural, skipped, failed)
	return err
}

const (
	stateDoc      = 0
	stateCmdStart = 1
	stateCmdCont  = 2
	stateExp      = 3
)

type Reader interface {
	Read(*Test) error
}

type testReader struct {
	scanner *lookaheadScanner
	state   int
}

func NewReader(r io.Reader) Reader {
	return &testReader{
		scanner: &lookaheadScanner{Scanner: bufio.NewScanner(r)},
		state:   stateDoc,
	}
}

func synErr(line int, msg string) error {
	return fmt.Errorf("syntax error parsing line %d: %s", line, msg)
}

type lookaheadScanner struct {
	*bufio.Scanner
	last   []byte
	unread []byte
}

func (l *lookaheadScanner) Scan() bool {
	if l.unread != nil {
		return true
	}

	return l.Scanner.Scan()
}

func (l *lookaheadScanner) Bytes() []byte {
	if l.unread != nil {
		unread := l.unread
		l.unread = nil
		return unread
	}
	l.last = l.Scanner.Bytes()
	return l.last
}

func (l *lookaheadScanner) Unread() {
	if l.unread != nil {
		panic("double Unread()")
	}
	l.unread = l.last
}

func (t *testReader) Read(test *Test) error {
	*test = Test{}

	i := 0

	for t.scanner.Scan() {
		buf := t.scanner.Bytes()
		line := make([]byte, len(buf))
		copy(line, buf) // buf's data gets stomped next iteration
		i++
		if len(line) == 0 {
			if t.state == stateDoc {
				test.doc = append(test.doc, line)
				continue
			}
			t.state = stateDoc
			if len(test.command) > 0 {
				t.scanner.Unread()
				return nil
			}
			continue
		}
		for {
			switch t.state {
			case stateDoc:
				if bytes.HasPrefix(line, []byte("  ")) {
					if bytes.HasPrefix(line, []byte("  $ ")) {
						t.state = stateCmdStart
						continue
					}
					return synErr(i, "expected '$ ' after two spaces")
				}
				test.doc = append(test.doc, line)
			case stateCmdStart:
				if len(line) < 5 {
					return synErr(i, "line too short")
				}
				// Assume next line is continuation; next state will
				// unread and go straight to exp state if necessary.
				t.state = stateCmdCont
				test.command = append(test.command, line[4:])
			case stateCmdCont:
				if !bytes.HasPrefix(line, []byte("  > ")) {
					t.state = stateExp
					continue
				}
				test.command = append(test.command, line[4:])
			case stateExp:
				if bytes.HasPrefix(line, []byte("  $ ")) {
					t.state = stateCmdStart
					t.scanner.Unread()
					return nil
				}
				if bytes.HasPrefix(line, []byte("  ")) {
					test.expResults = append(test.expResults, line[2:])
				} else {
					t.state = stateDoc
					t.scanner.Unread()
					return nil
				}
			}
			break
		}
	}
	return io.EOF
}

// Escape unprintable characters in a string, if there are any.
//
// If any character was escaped, append (esc) keyword to the end of the output string.
func escape(s []byte) string {
	a := string(s)
	b := strconv.Quote(a)
	b = b[1 : len(b)-1]
	if a != b {
		b += " (esc)"
	}
	return b
}

func byteSlicesToString(slice [][]byte) string {
	return string(bytes.Join(slice, []byte("\n")))
}
