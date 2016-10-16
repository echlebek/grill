package grill

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
LANGAUGE=C
TZ=GMT
COLUMNS=80
CDPATH=''
GREP_OPTIONS=''`

// DefaultTestContext creates a new TestContext with environment defaults.
func DefaultTestContext(testdir, shell string, stdout, stderr io.Writer) (TestContext, error) {
	// TODO support TESTFILE elsewhere
	td, err := ioutil.TempDir("", "grill")
	env := []string{
		fmt.Sprintf("TMPDIR=%s", td),
		fmt.Sprintf("TEMP=%s", td),
		fmt.Sprintf("TMP=%s", td),
		fmt.Sprintf("GRILLTMP=%s", td),
		fmt.Sprintf("TESTDIR=%s", testdir),
		fmt.Sprintf("TESTSHELL=%q", shell),
		"LANG=C",
		"LC_ALL=C",
		"LANGAUGE=C",
		"TZ=GMT",
		"COLUMNS=80",
		"CDPATH=''",
		"GREP_OPTIONS=''",
	}
	return TestContext{
		Shell:   strings.Split(shell, " "),
		WorkDir: td,
		Environ: env,
		Stdout:  stdout,
		Stderr:  stderr,
	}, err
}

// Cleanup removes the working directory of the test.
func (t TestContext) Cleanup() error {
	return os.RemoveAll(t.WorkDir)
}

// Run runs t within the TestContext. An error is returned if there is an
// error in executing the test.
func (t *Test) Run(ctx TestContext) error {
	buf := new(bytes.Buffer)
	if len(t.command) < 1 {
		return errors.New("empty command")
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

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("couldn't run command: %s", err)
	}
	if err = cmd.Wait(); err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			buf.Write(exitErr.Stderr)
			status := exitErr.Sys()
			if s, ok := status.(syscall.WaitStatus); ok {
				fmt.Fprintf(buf, "[%d]", s.ExitStatus())
			}
			err = nil
		}
	}

	t.obsResults = bytes.Split(buf.Bytes(), []byte{'\n'})
	if len(t.obsResults[len(t.obsResults)-1]) == 0 {
		t.obsResults = t.obsResults[:len(t.obsResults)-1]
	}

	if t.Failed() {
		if _, err := ctx.Stdout.Write([]byte{'!'}); err != nil {
			log.Println(err)
		}
	} else {
		if _, err := ctx.Stdout.Write([]byte{'.'}); err != nil {
			log.Println(err)
		}
	}

	return err
}
