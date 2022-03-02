package grill

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestRunSuite(t *testing.T) {
	suite := &TestSuite{
		Tests: []Test{
			{
				doc: [][]byte{
					[]byte("This is a test"),
				},
				command: [][]byte{[]byte("echo foobar")},
				expResults: [][]byte{
					[]byte("foobar"),
				},
			},
		},
	}

	stdout := new(bytes.Buffer)

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		r := recover()
		os.RemoveAll(dir)
		if r != nil {
			panic(r)
		}
	}()

	ctx := TestContext{
		Stdout:  stdout,
		Shell:   []string{"bash"},
		WorkDir: dir,
		Environ: os.Environ(),
	}

	if err := suite.Run(ctx); err != nil {
		t.Fatal(err)
	}

	test := &suite.Tests[0]

	join := byteSlicesToString
	if got, want := join(test.obsResults), join(test.expResults); got != want {
		t.Errorf("bad test output: got %q, want %q", got, want)
	}
}
