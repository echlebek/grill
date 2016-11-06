package main

import (
	"errors"
	"flag"
)

var (
	opts = struct {
		version     *bool
		quiet       *bool
		verbose     *bool
		interactive *bool
		debug       *bool
		yes         *bool
		no          *bool
		preserveEnv *bool
		keepTmpdir  *bool
		shell       *string
		shellOpts   *string
		xunitFile   *string
		indent      *int
	}{
		version:     flag.Bool("version", false, "show version information and exit"),
		quiet:       flag.Bool("quiet", false, "don't print diffs"),
		verbose:     flag.Bool("verbose", false, "show filenames and test status"),
		interactive: flag.Bool("interactive", false, "interactively merge changed test output"),
		debug:       flag.Bool("debug", false, "write script output directly to the terminal"),
		yes:         flag.Bool("yes", false, "answer yes to all questions"),
		no:          flag.Bool("no", false, "answer no to all questions"),
		preserveEnv: flag.Bool("preserve-env", false, "don't reset common environment variables"),
		keepTmpdir:  flag.Bool("keep-tmpdir", false, "keep temporary directories"),
		shell:       flag.String("shell", "/bin/sh", "shell to use for running tests"),
		shellOpts:   flag.String("shell-opts", "", "arguments to invoke shell with"),
		xunitFile:   flag.String("xunit-file", "", "path to write xUnit XML output"),
		indent:      flag.Int("indent", 2, "number of spaces to use for indentation"),
	}
)

func validateOptions() error {
	if *opts.yes && *opts.no {
		return errors.New("use of mutually exclusive -yes and -no")
	}
	if *opts.indent < 1 {
		return errors.New("-indent must be >= 1")
	}
	return nil
}
