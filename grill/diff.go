package grill

import (
	"bytes"
	"path/filepath"
	"regexp"

	"github.com/mb0/diff"
)

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

func Diff(a, b []byte) []byte {
	// TODO: Add diff header and context
	alines := bytes.Split(a, []byte("\n"))
	blines := bytes.Split(b, []byte("\n"))
	result := make([]byte, 0, len(alines)+len(blines))

	d := FuzzyMatchData{alines, blines}

	changes := diff.Diff(len(d.a), len(d.b), d)

	for _, c := range changes {
		delLines := alines[c.A : c.A+c.Del]
		insLines := blines[c.B : c.B+c.Ins]

		for _, line := range delLines {
			result = append(result, '-', ' ')
			result = append(result, line...)
			result = append(result, '\n')
		}
		for _, line := range insLines {
			result = append(result, '+', ' ')
			result = append(result, line...)
			result = append(result, '\n')
		}
	}
	return result
}
