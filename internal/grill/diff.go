package grill

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"

	"github.com/echlebek/diff"
	"github.com/echlebek/glob"
)

// TODO so far unused
const ContextLines = 5

// DiffData contains output difference data for a single test case.
type DiffData struct {
	a       [][]byte
	b       [][]byte
	changes []diff.Change
}

func (f DiffData) Equal(i, j int) bool {
	a := f.a[i]
	b := f.b[j]

	var v bool

	if bytes.HasSuffix(a, []byte(" (re)")) {
		v = matchRegexp(a[:len(a)-5], b)
	}
	if bytes.HasSuffix(a, []byte(" (glob)")) {
		v = matchGlob(a[:len(a)-7], b)
	}
	if bytes.HasSuffix(a, []byte(" (esc)")) {
		v = matchEsc(a[:len(a)-6], b)
	}

	// All of the keywords may appear verbatim in command
	// output, so check for direct equality every time.
	return v || bytes.Equal(a, b)
}

func matchRegexp(a, b []byte) bool {
	if len(a) == 0 {
		// Regex cannot be empty
		return false
	}
	match, err := regexp.Match(string(a), b)
	if err != nil {
		return bytes.Equal(a, b)
	}
	return match
}

func matchGlob(a, b []byte) bool {
	if len(a) == 0 {
		// Glob cannot be empty
		return false
	}
	match, err := glob.Match(string(a), string(b))
	if err != nil {
		return bytes.Equal(a, b)
	}
	return match
}

func matchEsc(a, b []byte) bool {
	// TODO there's probably a cleaner way to do it
	s, err := strconv.Unquote(`"` + string(a) + `"`)
	if err != nil {
		return false
	}
	return s == string(b)
}

// NewDiff computes difference between two slices of text lines.
func NewDiff(a, b []byte) DiffData {
	alines := bytes.Split(a, []byte("\n"))
	blines := bytes.Split(b, []byte("\n"))

	d := DiffData{a: alines, b: blines}
	d.changes = diff.Diff(len(d.a), len(d.b), d)

	return d
}

// ToString formats the diff data into a unified diff.
func (d DiffData) ToString(name string) []byte {
	if len(d.changes) == 0 {
		return nil
	}

	w := new(bytes.Buffer)

	fmt.Fprint(w, "--- ", name, "\n")
	fmt.Fprint(w, "+++ ", name, ".err", "\n")

	for _, c := range d.changes {
		fmt.Fprintf(w, "@@ -%d,%d +%d,%d @@\n", c.A+1, c.Del+1, c.B+1, c.Ins+1)

		delLines := d.a[c.A : c.A+c.Del]
		insLines := d.b[c.B : c.B+c.Ins]

		for _, line := range delLines {
			fmt.Fprint(w, "-  ", string(line), "\n")
		}
		for _, line := range insLines {
			fmt.Fprint(w, "+  ", string(line), "\n")
		}
	}
	return w.Bytes()
}

// Write writes diff in the unified text format.
//
// aLineNo and bLineNo set the initial line numbers, which is useful
// when there's more than one diff in the file. If there's a single
// diff in the file, then both are set to 1.
func (d DiffData) Write(w io.Writer, aLineNo int, bLineNo int) error {
	for _, c := range d.changes {
		aDiff := 0
		if c.Del == 0 {
			aDiff -= 1
		}

		bDiff := 0
		if c.Ins == 0 {
			bDiff -= 1
		}

		fmt.Fprintf(w, "@@ -%d,%d +%d,%d @@\n",
			c.A+aLineNo+aDiff, c.Del,
			c.B+bLineNo+bDiff, c.Ins)

		delLines := d.a[c.A : c.A+c.Del]
		insLines := d.b[c.B : c.B+c.Ins]

		for _, line := range delLines {
			if _, err := fmt.Fprint(w, "-  ", string(line), "\n"); err != nil {
				return err
			}
		}
		for _, line := range insLines {
			if _, err := fmt.Fprint(w, "+  ", escape(line), "\n"); err != nil {
				return err
			}
		}
	}
	return nil
}

// WriteDiff writes suite diff in the unified text format.
func WriteDiff(w io.Writer, suite *TestSuite) error {
	expLineNo := 1
	obsLineNo := 1

	if _, err := fmt.Fprint(w, "--- ", suite.Name, "\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, "+++ ", suite.Name, ".err", "\n"); err != nil {
		return err
	}

	for _, t := range suite.Tests {
		d := t.diff

		expOutLineNo := expLineNo + len(t.doc) + len(t.command)
		obsOutLineNo := obsLineNo + len(t.doc) + len(t.command)

		if err := d.Write(w, expOutLineNo, obsOutLineNo); err != nil {
			return err
		}

		expLineNo = expOutLineNo + len(t.expResults)
		obsLineNo = obsOutLineNo + len(t.obsResults)
	}
	return nil
}
