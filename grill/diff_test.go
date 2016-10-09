package grill

import "testing"

type test struct {
	Old          string
	New          string
	ExpectedDiff string
}

var tests = []test{
	test{
		Old: `Here is a line
there are many like it
but this one is mine.`,
		New: `Here is a mine
there are many like it
but this one is mine.
`,
		ExpectedDiff: `- Here is a line
+ Here is a mine
+ 
`,
	},
	test{
		Old: `Here is a line
There are \d+ like it (re)
But this one is mine.`,
		New: `Here is a line
There are 37 like it
But this one is mine.`,
		ExpectedDiff: ``,
	},
	test{
		Old: `Here is a line
There are to* like it (glob)
But this one is mine.`,
		New: `Here is a line
There are tons like it
But this one is mine.`,
		ExpectedDiff: ``,
	},
	test{
		Old: `Here is some text
The next few lines
will change quite a bit
especially this one
but not this one.`,
		New: `Here is some text
Blah blah blah
Foo bar baz
I like pizza
Check out our great deals on ink and toner
but not this one.`,
		ExpectedDiff: `- The next few lines
- will change quite a bit
- especially this one
+ Blah blah blah
+ Foo bar baz
+ I like pizza
+ Check out our great deals on ink and toner
`,
	},
}

func TestDiff(t *testing.T) {
	for i, test := range tests {
		got := string(Diff([]byte(test.Old), []byte(test.New)))
		want := test.ExpectedDiff
		if got != want {
			t.Errorf("test %d: got %q, want %q", i, got, want)
		}
	}
}
