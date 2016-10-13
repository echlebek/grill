package grill

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
)

var ErrNoCommand = errors.New("couldn't read any commands")

// Test is a single grill test. It is comprised of documentation, commands, and
// expected test results.
type Test struct {
	doc        [][]byte
	command    []byte
	expResults [][]byte
	obsResults [][]byte
}

func (t Test) Doc() string {
	return byteSlicesToString(t.doc)
}

func (t Test) Command() []string {
	parts := bytes.Split(t.command, []byte{' '})
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) > 0 {
			result = append(result, string(p))
		}
	}
	return result
}

func byteSlicesToString(slice [][]byte) string {
	if len(slice) == 0 {
		return ""
	}
	result := make([]byte, 0)
	for _, b := range slice {
		result = append(result, b...)
		result = append(result, '\n')
	}
	return string(result[:len(result)-1])
}

func (t Test) ExpectedResults() string {
	return byteSlicesToString(t.expResults)
}

func (t Test) ObservedResults() string {
	return byteSlicesToString(t.obsResults)
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
		if _, err := fmt.Fprintf(f, "  $ %s\n", string(t.command)); err != nil {
			return fmt.Errorf("couldn't write %s: %s", tErr, err)
		}
		for _, e := range t.obsResults {
			if _, err := fmt.Fprintf(f, "  %s\n", e); err != nil {
				return fmt.Errorf("couldn't write %s: %s", tErr, err)
			}
		}
	}
	return nil
}

// WriteReport writes out a report of the diff between the test's
// ExpectedResults and ObservedResults, or a '.' if the testsuite
// succeeded.
func (suite TestSuite) WriteReport(w io.Writer) error {
	if _, err := w.Write([]byte{'\n'}); err != nil {
		return err
	}
	tests, failed := 0, 0
	for _, t := range suite.Tests {
		if t.Failed() {
			exp, obs := t.ExpectedResults(), t.ObservedResults()
			diff := Diff([]byte(exp), []byte(obs), suite.Name)
			if _, err := w.Write(diff); err != nil {
				return fmt.Errorf("couldn't write %q: %s", suite.Name+".err", err)
			}
			failed++
		}
		tests++
	}
	plural := "s"
	if tests == 1 {
		plural = ""
	}
	_, err := fmt.Fprintf(w, "# Ran %d test%s, %d failed.\n", tests, plural, failed)
	return err
}

func (t *Test) Failed() bool {
	return t.ExpectedResults() != t.ObservedResults()
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
				if bytes.HasSuffix(line, []byte("\\")) {
					t.state = stateCmdCont
				} else {
					t.state = stateExp
				}
				test.command = append(test.command, bytes.Trim(line[4:], "\n")...)
			case stateCmdCont:
				if !bytes.HasPrefix(line, []byte("  > ")) {
					return synErr(i, "truncated command")
				}
				if len(line) < 5 {
					return synErr(i, "line too short")
				}
				trimmed := line[4:]
				if bytes.HasSuffix(line, []byte("\\")) {
					trimmed = trimmed[:len(trimmed)-1]
				} else {
					t.state = stateExp
				}
				args := bytes.Split(trimmed, []byte(" "))
				for _, a := range args {
					if len(a) > 0 {
						test.command = append(test.command, bytes.Trim(a, "\n")...)
					}
				}
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
