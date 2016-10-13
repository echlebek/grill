package grill

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestRunTest(t *testing.T) {
	test := &Test{
		doc: [][]byte{
			[]byte("This is a test"),
		},
		command: []byte("echo foobar"),
		expResults: [][]byte{
			[]byte("foobar"),
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
		Shell:   []string{},
		WorkDir: dir,
		Environ: os.Environ(),
	}

	if err := test.Run(ctx); err != nil {
		t.Fatal(err)
	}

	if got, want := string(stdout.Bytes()), "."; got != want {
		t.Errorf("bad test status output: got %q, want %q", got, want)
	}

	if got, want := test.ExpectedResults(), test.ObservedResults(); got != want {
		t.Errorf("bad test output: got %q, want %q", got, want)
	}

}
