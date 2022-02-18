package grill

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/mb0/diff"
	"github.com/mb0/glob"
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

	if bytes.HasSuffix(a, []byte(" (re)")) {
		return matchRegexp(a[:len(a)-5], b)
	}
	if bytes.HasSuffix(a, []byte(" (glob)")) {
		return matchGlob(a[:len(a)-7], b)
	}
	return bytes.Equal(a, b)
}

func matchRegexp(a, b []byte) bool {
	match, err := regexp.Match(string(a), b)
	if err != nil {
		return bytes.Equal(a, b)
	}
	return match
}

func matchGlob(a, b []byte) bool {
	match, err := glob.Match(string(a), string(b))
	if err != nil {
		return bytes.Equal(a, b)
	}
	return match
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