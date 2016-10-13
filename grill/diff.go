package grill

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/mb0/diff"
)

const ContextLines = 5

type FuzzyMatchData struct {
	a [][]byte
	b [][]byte
}

func (f FuzzyMatchData) Equal(i, j int) bool {
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
	match, err := filepath.Match(string(a), string(b))
	if err != nil {
		return bytes.Equal(a, b)
	}
	return match
}

func Diff(a, b []byte, name string) []byte {
	alines := bytes.Split(a, []byte("\n"))
	blines := bytes.Split(b, []byte("\n"))
	w := new(bytes.Buffer)

	d := FuzzyMatchData{alines, blines}

	changes := diff.Diff(len(d.a), len(d.b), d)
	if len(changes) == 0 {
		return nil
	}
	fmt.Fprint(w, "--- ", name, "\n")
	fmt.Fprint(w, "+++ ", name, ".err", "\n")

	for _, c := range changes {
		fmt.Fprintf(w, "@@ -%d,%d +%d,%d @@\n", c.A+1, c.Del+1, c.B+1, c.Ins+1)

		delLines := alines[c.A : c.A+c.Del]
		insLines := blines[c.B : c.B+c.Ins]

		for _, line := range delLines {
			fmt.Fprint(w, "-  ", string(line), "\n")
		}
		for _, line := range insLines {
			fmt.Fprint(w, "+  ", string(line), "\n")
		}
	}
	return w.Bytes()
}
