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

// Test is a single cram test. It is comprised of documentation, commands, and
// expected test results.
// TODO: Use [][]byte for ExpectedResults to avoid copies.
type Test struct {
	doc        [][]byte
	command    []byte
	expResults [][]byte
	obsResults [][]byte
}

func (t Test) Doc() io.Reader {
	return multiReader(t.doc)
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

func (t Test) ExpectedResults() io.Reader {
	return multiReader(t.expResults)
}

func (t Test) ObservedResults() io.Reader {
	return multiReader(t.obsResults)
}

// I think this is a little silly, but it reflects what I'd like
// to see the interface look like.
func multiReader(b [][]byte) io.Reader {
	readers := make([]io.Reader, len(b)*2)
	for i := range readers {
		if i%2 == 0 {
			readers[i] = bytes.NewReader(b[i/2])
		} else {
			readers[i] = bytes.NewReader([]byte{'\n'})
		}
	}
	return io.MultiReader(readers...)
}

// A TestSuite represents a single cram test file.
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
func (suite TestSuite) WriteReport(w io.Writer, differ Differ) error {
	if !suite.Failed() {
		if _, err := w.Write([]byte{'.'}); err != nil {
			return fmt.Errorf("couldn't write %q: %s", suite.Name+".err", err)
		}
	}
	for _, t := range suite.Tests {
		exp, obs := t.ExpectedResults(), t.ObservedResults()
		diff := differ.Diff(exp, obs, suite.Name, suite.Name+".err")
		if _, err := w.Write(diff); err != nil {
			return fmt.Errorf("couldn't write %q: %s", suite.Name+".err", err)
		}
	}
	return nil
}

func (t *Test) Failed() bool {
	if len(t.expResults) != len(t.obsResults) {
		return true
	}
	for i := range t.expResults {
		if !bytes.Equal(t.expResults[i], t.obsResults[i]) {
			return true
		}
	}
	return false
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
