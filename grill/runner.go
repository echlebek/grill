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
	// TODO Handle --preserve-env flag
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

	// Add test specific variables
	testdir, err := filepath.Abs(filepath.Dir(t.Filepath))
	if err != nil {
		return err
	}

	// Create working directory for individual source file
	cmd.Dir = filepath.Join(ctx.WorkDir, filepath.Base(t.Filepath))
	if err := os.Mkdir(cmd.Dir, 0700); err != nil && !os.IsExist(err) {
		return err
	}

	cmd.Env = append(ctx.Environ, []string{
		// TODO escape spaces in paths?
		fmt.Sprintf("TESTFILE=%s", t.Filepath),
		fmt.Sprintf("TESTDIR=%s", testdir),
	}...)

	if err = cmd.Start(); err != nil {
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

	t.diff = NewDiff([]byte(t.ExpectedResults()), []byte(t.ObservedResults()))

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
