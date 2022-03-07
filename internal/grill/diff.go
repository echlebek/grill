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

// DiffData contains data for computing difference between two blocks of lines.
type DiffData struct {
	a [][]byte
	b [][]byte
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

// Diff computes change between expected sequence of
// lines (a) and observed sequence of lines (b)
func Diff(a, b [][]byte) []*Change {
	d := DiffData{a: a, b: b}
	var changes []*Change

	for _, c := range diff.Diff(len(d.a), len(d.b), d) {
		changes = append(changes, &Change{
			A:   c.A,
			B:   c.B,
			Del: c.Del,
			Ins: c.Ins,
		})
	}

	return changes
}

// Change is a single block of differences between two sequences of lines.
//
// A/Del indexes the expected lines and B/Ins indexes the observed lines.
//
// Change that represents diff context with unchanged lines has zero Del
// and Ins counts and non-zero Same count.
type Change struct {
	A    int
	B    int
	Del  int
	Ins  int
	Same int
}

// Hunk is a continuous sequence of changes with overlapping contexts.
//
// Context blocks are represented by changes where len(Same)>0 and are
// inserted automatically when new changes are appendeed and when
// hunk is finalized.
type Hunk struct {
	changes []*Change
	ctxLen  int
}

// NewHunk creates a new hunk with a single change.
func NewHunk(c *Change, ctxLen int) *Hunk {
	// First change; insert leading context
	// Populate B since it's necessary to produce header.
	same := intMin(c.A, ctxLen)
	if same != intMin(c.B, ctxLen) {
		panic("before/after diff offsets don't match")
	}
	ctx := &Change{
		A:    intMax(0, c.A-ctxLen),
		B:    intMax(0, c.B-ctxLen),
		Same: same,
	}
	return &Hunk{
		changes: []*Change{ctx, c},
		ctxLen:  ctxLen,
	}
}

// AppendChange appends new change to the end of the hunk.
func (h *Hunk) AppendChange(c *Change) {
	// Context between last change and this one
	prev := h.changes[len(h.changes)-1]
	if prev.Same > 0 {
		panic("hunk has already been finalized")
	}

	start := prev.A + prev.Del
	ctx := &Change{
		A:    start,
		Same: c.A - start,
	}
	h.changes = append(h.changes, ctx, c)
}

// Finalize marks hunk as complete.
//
// After hunk is finalized, no more changes can be added.
func (h *Hunk) Finalize(numLinesA int) {
	// Add tail context
	prev := h.changes[len(h.changes)-1]
	start := prev.A + prev.Del
	ctx := &Change{
		A:    start,
		Same: intMin(h.ctxLen, numLinesA-start),
	}
	h.changes = append(h.changes, ctx)
}

// Write writes hunk in a unified diff format.
func (h *Hunk) Write(w io.Writer, linesA [][]byte, linesB [][]byte) error {
	numDel, numIns := 0, 0
	for _, c := range h.changes {
		numDel += c.Del + c.Same
		numIns += c.Ins + c.Same
	}

	dA := 0
	if numDel == 0 {
		dA -= 1
	}

	dB := 0
	if numIns == 0 {
		dB -= 1
	}

	lead := h.changes[0]

	_, err := fmt.Fprintf(w, "@@ -%d,%d +%d,%d @@\n", lead.A+dA+1, numDel, lead.B+dB+1, numIns)
	if err != nil {
		return err
	}

	for _, c := range h.changes {
		if c.Same > 0 {
			for _, line := range linesA[c.A : c.A+c.Same] {
				if _, err := fmt.Fprint(w, " ", string(line), "\n"); err != nil {
					return err
				}
			}
		} else {
			for _, line := range linesA[c.A : c.A+c.Del] {
				if _, err := fmt.Fprint(w, "-", string(line), "\n"); err != nil {
					return err
				}
			}
			for _, line := range linesB[c.B : c.B+c.Ins] {
				if _, err := fmt.Fprint(w, "+", escape(line), "\n"); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// CreateHunks creates a sequence of hunks from a sequence of changes.
//
// aLen is the total number of expected (old) lines. ctxLen is the number
// of context lines to print before/after each change.
//
// If contexts of a of adjacent changes overlap, then the following
// change is merged into the preceding one.
func CreateHunks(changes []*Change, aLen int, ctxLen int) []*Hunk {
	var hunks []*Hunk
	var h *Hunk
	var prev *Change

	for _, c := range changes {
		if len(hunks) == 0 || c.A-(prev.A+prev.Del) > 2*ctxLen {
			h = NewHunk(c, ctxLen)
			hunks = append(hunks, h)
		} else {
			h.AppendChange(c)
		}
		prev = c
	}

	for _, h := range hunks {
		h.Finalize(aLen)
	}

	return hunks
}

// WriteDiff writes suite diff in the unified text format.
func (suite *TestSuite) WriteDiff(w io.Writer, ctxLen int) error {
	var expLines [][]byte
	var obsLines [][]byte
	var changes []*Change

	for _, t := range suite.Tests {
		var cmdLines [][]byte
		for i, line := range t.command {
			if i == 0 {
				cmdLines = append(cmdLines, append([]byte("  $ "), line...))
			} else {
				cmdLines = append(cmdLines, append([]byte("  > "), line...))
			}
		}

		expLines = append(expLines, t.doc...)
		expLines = append(expLines, cmdLines...)

		obsLines = append(obsLines, t.doc...)
		obsLines = append(obsLines, cmdLines...)

		for _, c := range t.changes {
			// Convert to absolute offsets.
			changes = append(changes, &Change{
				A:   c.A + len(expLines),
				B:   c.B + len(obsLines),
				Del: c.Del,
				Ins: c.Ins,
			})
		}

		for _, line := range t.expResults {
			expLines = append(expLines, append([]byte("  "), line...))
		}
		for _, line := range t.obsResults {
			obsLines = append(obsLines, append([]byte("  "), line...))
		}
	}

	hunks := CreateHunks(changes, len(expLines), ctxLen)

	if _, err := fmt.Fprint(w, "--- ", suite.Name, "\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, "+++ ", suite.Name, ".err", "\n"); err != nil {
		return err
	}

	for _, h := range hunks {
		if err := h.Write(w, expLines, obsLines); err != nil {
			return err
		}
	}

	return nil
}

func intMin(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func intMax(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
