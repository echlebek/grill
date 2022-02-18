package internal

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"
)

const test1 = `Run grill examples:

  $ grill -q examples examples/fail.t
  .s.!.s.
  # Ran 7 tests, 2 skipped, 1 failed.
  [1]
  $ md5 examples/fail.t examples/fail.t.err
  .*\b0f598c2b7b8ca5bcb8880e492ff6b452\b.* (re)
  .*\b7a23dfa85773c77648f619ad0f9df554\b.* (re)
  $ rm examples/fail.t.err`

func makeSpecs() []spec {
	return []spec{
		{
			doc:     "Run grill examples:\n",
			command: []byte("grill -q examples examples/fail.t"),
			results: ".s.!.s.\n# Ran 7 tests, 2 skipped, 1 failed.\n[1]",
		},
		{
			command: []byte("md5 examples/fail.t examples/fail.t.err"),
			results: ".*\\b0f598c2b7b8ca5bcb8880e492ff6b452\\b.* (re)\n.*\\b7a23dfa85773c77648f619ad0f9df554\\b.* (re)",
		},
		{
			command: []byte("rm examples/fail.t.err"),
		},
	}
}

type spec struct {
	doc     string
	command []byte
	results string
}

func TestReadTests(t *testing.T) {
	t.Parallel()
	specs := makeSpecs()
	buf := new(bytes.Buffer)
	if _, err := buf.Write([]byte(test1)); err != nil {
		t.Fatal(err)
	}

	r := NewReader(bytes.NewReader(buf.Bytes()))

	var (
		test Test
		err  error
	)

	var tests []Test

	for err != io.EOF && len(tests) < 4 {
		err = r.Read(&test)
		if err != nil && err != io.EOF {
			t.Fatal(err)
		}
		tests = append(tests, test)
	}

	if len(tests) != len(specs) {
		t.Fatalf("wrong number of tests: got %d, want %d", len(tests), len(specs))
	}

	for i, spec := range specs {
		test := tests[i]
		if got, want := test.Doc(), spec.doc; got != want {
			t.Errorf("test %d: bad doc: got %q, want %q", i, got, want)
		}
		if !reflect.DeepEqual(spec.command, test.command) {
			t.Errorf("test %d: bad cmd: got %q, want %q", i, fmt.Sprint(test.Command()), spec.command)
		}
		if got, want := test.ExpectedResults(), spec.results; got != want {
			t.Errorf("test %d: bad expected results: got %q, want %q", i, got, want)
		}
	}
}
