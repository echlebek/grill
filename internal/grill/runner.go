package grill

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// TestContext specifies an execution environment for running a test.
type TestContext struct {
	Environ []string
	WorkDir string
	Shell   []string
	Stdout  io.Writer
	Stderr  io.Writer
}

// Default environment variables set by grill.
const DefaultEnvironment = `LANG=C
LC_ALL=C
LANGUAGE=C
TZ=GMT
COLUMNS=80
CDPATH=
GREP_OPTIONS=`

// DefaultTestContext creates a new TestContext with environment defaults.
//
// The function is meant to be called once per grill command invocation and
// creates two things:
//
//  - An overall working directory root in default TMPDIR
//  - A single local {workdir}/tmp temporary directory for all of the executed tests
//
// As tests execute later on, they will create named sub-directories
// that will serve as their individual working directories.
func DefaultTestContext(shell string, stdout, stderr io.Writer) (TestContext, error) {
	wd, err := ioutil.TempDir("", "grilltests")
	td := filepath.Join(wd, "tmp")
	if err := os.Mkdir(td, 0700); err != nil {
		return TestContext{}, err
	}

	env := []string{
		fmt.Sprintf("TMPDIR=%s", td),
		fmt.Sprintf("TEMP=%s", td),
		fmt.Sprintf("TMP=%s", td),
		fmt.Sprintf("GRILLTMP=%s", td),
		fmt.Sprintf("CRAMTMP=%s", td),
		fmt.Sprintf("TESTSHELL=%q", shell),
	}
	env = append(env, strings.Split(DefaultEnvironment, "\n")...)
	env = append(env, os.Environ()...)
	return TestContext{
		Shell:   strings.Split(shell, " "),
		WorkDir: wd,
		Environ: env,
		Stdout:  stdout,
		Stderr:  stderr,
	}, err
}

// Cleanup removes the working directory of the test.
func (t TestContext) Cleanup() error {
	return os.RemoveAll(t.WorkDir)
}

// Run runs the test within the TestContext. An non-nil error
// indicates a failure to run the test. Use Test.Failed() to find
// out if the test ran but did not produce the expected output.
func (t *Test) Run(ctx TestContext) error {
	buf := new(bytes.Buffer)
	if len(t.command) < 1 {
		// No command, will be considered skipped
		return nil
	}

	var cdr []string
	if len(ctx.Shell) > 1 {
		cdr = ctx.Shell[1:]
	}
	cmd := exec.Command(ctx.Shell[0], cdr...)
	cmd.Stdout = buf
	cmd.Stderr = buf
	cmd.Stdin = t.Command()
	cmd.Env = ctx.Environ
	cmd.Dir = ctx.WorkDir

	var err error

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("couldn't run command: %s", err)
	}

	if err = cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			buf.Write(exitErr.Stderr)
			status := exitErr.Sys()
			if s, ok := status.(syscall.WaitStatus); ok {
				fmt.Fprintf(buf, "[%d]", s.ExitStatus())
			}
			err = nil
		} else {
			panic(fmt.Sprintf("command exited with unexpected error: %s", err))
		}
	} else {
		b := buf.Bytes()
		if len(b) > 0 && b[len(b)-1] != '\n' {
			buf.WriteString(" (no-eol)")
		}
	}

	t.obsResults = bytes.Split(buf.Bytes(), []byte{'\n'})
	if len(t.obsResults[len(t.obsResults)-1]) == 0 {
		t.obsResults = t.obsResults[:len(t.obsResults)-1]
	}

	t.diff = NewDiff([]byte(t.ExpectedResults()), []byte(t.ObservedResults()))
	return err
}

// Run runs the entire suite within the TestContext. An non-nil error
// indicates a failure to run the test. Use TestSuite.Failed() to find
// out if the test ran but did not produce the expected output.
//
// At the end, Run prints suite status glyph to ctx.Stdout.
func (suite *TestSuite) Run(ctx TestContext) error {
	// Add test specific variables
	testdir, err := filepath.Abs(filepath.Dir(suite.Name))
	if err != nil {
		return err
	}

	// Update workdir for each individual test;
	// OK to set fields since ctx is passed by value.
	ctx.WorkDir = filepath.Join(ctx.WorkDir, strings.TrimPrefix(suite.Name, "/"))
	if err := os.MkdirAll(ctx.WorkDir, 0700); err != nil {
		return err
	}

	ctx.Environ = append(ctx.Environ, []string{
		// TODO escape spaces in paths?
		fmt.Sprintf("TESTFILE=%s", filepath.Base(suite.Name)),
		fmt.Sprintf("TESTDIR=%s", testdir),
	}...)

	for i := 0; i < len(suite.Tests); i++ {
		if err := suite.Tests[i].Run(ctx); err != nil {
			return err
		}
	}

	if _, err := ctx.Stdout.Write(suite.StatusGlyph()); err != nil {
		return err
	}

	return nil
}
