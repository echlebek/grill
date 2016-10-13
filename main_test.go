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
	if err := Main([]string{"grill", ctx.Test.Name()}, ctx.Stdout, ctx.Stderr); err != nil {
		t.Fatal(err)
	}
	if want, got := "!", string(ctx.Stdout.Bytes()); got != want {
		t.Errorf("bad stdout: got %q, want %q", got, want)
	}
	stderr := string(ctx.Stderr.Bytes())
	if !strings.HasSuffix(stderr, "@@ -1,2 +1,2 @@\n-  foobaz\n+  foobar\n# Ran 1 test, 1 failed.\n") {
		t.Errorf("bad Stderr: %q", stderr)
	}
}
