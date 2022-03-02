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

	// Temporary directory and paths for output / status
	// Will contain:
	//   * status - common status file, one byte per test exit status code; no newline
	//   * out.0, out.1, ... - output for each individual test command in a suite.
	cmdDir := ctx.WorkDir + ".cmd"
	if err := os.MkdirAll(cmdDir, 0700); err != nil {
		return err
	}

	outBasePath := filepath.Join(cmdDir, "out")
	statusPath := filepath.Join(cmdDir, "status")
	statusCmd := fmt.Sprintf("echo -n $? >>%s\n", statusPath)

	ctx.Environ = append(ctx.Environ, []string{
		// TODO escape spaces in paths?
		fmt.Sprintf("TESTFILE=%s", filepath.Base(suite.Name)),
		fmt.Sprintf("TESTDIR=%s", testdir),
	}...)

	// TODO use <testname>.cmd/in temporary file instead for easier debugging?
	script := new(bytes.Buffer)
	for i, t := range suite.Tests {
		// Redirect pipes to dedicated output file for each test command.
		// Write command status to a single status file.
		script.WriteString(fmt.Sprintf("exec >%s.%d 2>&1\n", outBasePath, i))
		for _, line := range t.command {
			script.Write(line)
			script.WriteByte('\n')
		}
		script.WriteString(statusCmd)
	}

	var shellOpts []string
	if len(ctx.Shell) > 1 {
		shellOpts = ctx.Shell[1:]
	}
	cmd := exec.Command(ctx.Shell[0], shellOpts...)
	cmd.Stdin = script
	cmd.Env = ctx.Environ
	cmd.Dir = ctx.WorkDir

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("couldn't run command: %s", err)
	}

	if err := cmd.Wait(); err != nil {
		// Last command in each script is the separator that echos return code
		// and so the script should always exit with zero. If it doesn't, then
		// it likely exited prematurely (e.g. developer had set -e in it)
		return fmt.Errorf("test exited with unexpected error: %s", err)
	}

	// Read the list of exit status codes
	status, err := os.ReadFile(statusPath)
	if err != nil {
		return fmt.Errorf("could not read test status: %s", err)
	}

	if len(status) != len(suite.Tests) {
		return fmt.Errorf("no. of status codes does not match no. of tests (%d != %d)",
			len(status), len(suite.Tests))
	}

	for i, _ := range suite.Tests {
		t := &suite.Tests[i]

		// Test output
		b, err := os.ReadFile(fmt.Sprintf("%s.%d", outBasePath, i))
		if err != nil {
			return fmt.Errorf("could not read test output: %s", err)
		}

		lines := bytes.Split(b, []byte{'\n'})
		if j := len(lines) - 1; len(lines[j]) != 0 {
			lines[j] = append(lines[j], []byte(" (no-eol)")...)
		} else {
			lines = lines[:j]
		}

		// Test exit status
		if s := status[i]; s != '0' {
			lines = append(lines, []byte{'[', s, ']'})
		}

		t.obsResults = lines
		t.diff = NewDiff(t.expResults, t.obsResults)
	}

	if _, err := ctx.Stdout.Write(suite.StatusGlyph()); err != nil {
		return err
	}

	return nil
}
