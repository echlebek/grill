package grill

import (
	"bytes"
	"testing"
)

type test struct {
	Old          string
	New          string
	ExpectedDiff string
}

var tests = []test{
	// No deletions, 1 insertion
	{
		Old: `there are many like it
but this one is mine.
`,
		New: `Here is a mine
there are many like it
but this one is mine.
`,
		ExpectedDiff: `@@ -0,0 +1,1 @@
+Here is a mine
`,
	},
	// 1 deletion, no insertions
	{
		Old: `Here is a line
there are many like it
but this one is mine.
`,
		New: `there are many like it
but this one is mine.
`,
		ExpectedDiff: `@@ -1,1 +0,0 @@
-Here is a line
`,
	},
	// 1 deletion, 1 insertion
	{
		Old: `Here is a line
there are many like it
but this one is mine.
`,
		New: `Here is a mine
there are many like it
but this one is mine.
`,
		ExpectedDiff: `@@ -1,1 +1,1 @@
-Here is a line
+Here is a mine
`,
	},
	// Regex match
	{
		Old: `Here is a line
There are \d+ like it (re)
But this one is mine.
`,
		New: `Here is a line
There are 37 like it
But this one is mine.
`,
		ExpectedDiff: ``,
	},
	// Glob match
	{
		Old: `Here is a line
There are to* like it (glob)
But this one is mine.
`,
		New: `Here is a line
There are tons like it
But this one is mine.
`,
		ExpectedDiff: ``,
	},
	// Multiple deletions and insertions
	{
		Old: `Here is some text
The next few lines
will change quite a bit
especially this one
but not this one.
`,
		New: `Here is some text
Blah blah blah
Foo bar baz
I like pizza
Check out our great deals on ink and toner
but not this one.
`,
		ExpectedDiff: `@@ -2,3 +2,4 @@
-The next few lines
-will change quite a bit
-especially this one
+Blah blah blah
+Foo bar baz
+I like pizza
+Check out our great deals on ink and toner
`,
	},
	// Multiple hunks
	{
		Old: `Here is some deleted text
The next few lines
will not change
at all
Here is some old text
except this one.
`,
		New: `The next few lines
will not change
Here is some added text
at all
Here is some new text
except this one.
`,
		ExpectedDiff: `@@ -1,1 +0,0 @@
-Here is some deleted text
@@ -3,0 +3,1 @@
+Here is some added text
@@ -5,1 +5,1 @@
-Here is some old text
+Here is some new text
`,
	},
}

func TestDiff(t *testing.T) {
	splitLines := func(b string) [][]byte {
		return bytes.Split([]byte(b), []byte("\n"))
	}

	for i, test := range tests {
		oldLines := splitLines(test.Old)
		newLines := splitLines(test.New)

		changes := Diff(oldLines, newLines)

		var b bytes.Buffer
		for _, h := range CreateHunks(changes, len(oldLines), 0) {
			_ = h.Write(&b, oldLines, newLines)
		}

		if got, want := b.String(), test.ExpectedDiff; got != want {
			t.Errorf("test %d: got %q, want %q", i, got, want)
		}
	}
}

var testsContext = []test{
	{
		// Full contexts on both sides
		Old: "a\nb\nd\ne",
		New: "a\nb\nc\nd\ne",
		ExpectedDiff: `@@ -1,4 +1,5 @@
 a
 b
+c
 d
 e
`,
	},
	{
		// Partial leading context
		Old: "b\nd\ne",
		New: "b\nc\nd\ne",
		ExpectedDiff: `@@ -1,3 +1,4 @@
 b
+c
 d
 e
`,
	},
	{
		// Partial trailing context
		Old: "a\nb\nd",
		New: "a\nb\nc\nd",
		ExpectedDiff: `@@ -1,3 +1,4 @@
 a
 b
+c
 d
`,
	},
	{
		// Two changes with overlapping contexts
		Old: "a\nc\nd\ne\nf",
		New: "a\nb\nc\nd\nf",
		ExpectedDiff: `@@ -1,5 +1,5 @@
 a
+b
 c
 d
-e
 f
`,
	},
}

func TestDiffContext(t *testing.T) {
	splitLines := func(b string) [][]byte {
		return bytes.Split([]byte(b), []byte("\n"))
	}

	for i, test := range testsContext {
		oldLines := splitLines(test.Old)
		newLines := splitLines(test.New)

		changes := Diff(oldLines, newLines)

		var b bytes.Buffer
		for _, h := range CreateHunks(changes, len(oldLines), 2) {
			_ = h.Write(&b, oldLines, newLines)
		}

		if got, want := b.String(), test.ExpectedDiff; got != want {
			t.Errorf("test %d: got %q, want %q", i, got, want)
		}
	}
}
