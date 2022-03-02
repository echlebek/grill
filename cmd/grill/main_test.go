package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testData = `Here is an example grill test

  $ echo foobar
  foobar
`
	failTestData = `Here is another example

  $ echo foobar
  foobaz
`
)

type testCtx struct {
	Dir    string
	Test   *os.File
	Stdout *bytes.Buffer
	Stderr *bytes.Buffer
}

func newTestCtx(test string) (testCtx, error) {
	var ctx testCtx
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return ctx, err
	}
	ctest, err := os.Create(filepath.Join(dir, "test.t"))
	if err != nil {
		defer os.RemoveAll(dir)
		return ctx, err
	}
	if _, err := ctest.Write([]byte(test)); err != nil {
		defer os.RemoveAll(dir)
		return ctx, err
	}

	ctx.Dir = dir
	ctx.Test = ctest
	ctx.Stdout = new(bytes.Buffer)
	ctx.Stderr = new(bytes.Buffer)

	return ctx, nil
}

func TestGrillFail(t *testing.T) {
	ctx, err := newTestCtx(failTestData)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := Main([]string{ctx.Test.Name()}, ctx.Stdout, ctx.Stderr), 1; got != want {
		t.Errorf("bad return code: got %d, want %d", got, want)
	}

	stdout := ctx.Stdout.String()
	if !strings.HasPrefix(stdout, "!\n") {
		t.Errorf("bad stdout: %q", stdout)
	}
	if !strings.HasSuffix(stdout, "@@ -4,1 +4,1 @@\n-  foobaz\n+  foobar\n# Ran 1 test, 0 skipped, 1 failed.\n") {
		t.Errorf("bad stdout: %q", stdout)
	}
}

func TestGrillPass(t *testing.T) {
	ctx, err := newTestCtx(testData)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := Main([]string{ctx.Test.Name()}, ctx.Stdout, ctx.Stderr), 0; got != want {
		t.Errorf("bad return code: got %d, want %d", got, want)
	}

	if want, got := ".\n# Ran 1 test, 0 skipped, 0 failed.\n", ctx.Stdout.String(); got != want {
		t.Errorf("bad stdout: got %q, want %q", got, want)
	}
}

func TestGrillVersion(t *testing.T) {
	ctx, err := newTestCtx(testData)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := Main([]string{"-version", "foo", "bar", "baz"}, ctx.Stdout, ctx.Stderr), 0; got != want {
		t.Errorf("bad return code: got %d, want %d", got, want)
	}
	got := ctx.Stderr.String()
	want := grillVersion + "\n"
	if got != want {
		t.Errorf("bad version: got %q, want %q", got, want)
	}
}
