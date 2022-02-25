package main

import (
	"errors"
	"flag"
)

var flags = flag.NewFlagSet("grill", flag.PanicOnError)

var opts = struct {
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
	version:     flags.Bool("version", false, "show version information and exit"),
	quiet:       flags.Bool("quiet", false, "don't print diffs"),
	verbose:     flags.Bool("verbose", false, "show filenames and test status (unsupported)"),
	interactive: flags.Bool("interactive", false, "interactively merge changed test output (unsupported)"),
	debug:       flags.Bool("debug", false, "write script output directly to the terminal (unsupported)"),
	yes:         flags.Bool("yes", false, "answer yes to all questions (unsupported)"),
	no:          flags.Bool("no", false, "answer no to all questions (unsupported)"),
	preserveEnv: flags.Bool("preserve-env", false, "don't reset common environment variables"),
	keepTmpdir:  flags.Bool("keep-tmpdir", false, "keep temporary directories"),
	shell:       flags.String("shell", "/bin/sh", "shell to use for running tests"),
	shellOpts:   flags.String("shell-opts", "", "arguments to invoke shell with (unsupported)"),
	xunitFile:   flags.String("xunit-file", "", "path to write xUnit XML output (unsupported)"),
	indent:      flags.Int("indent", 2, "number of spaces to use for indentation (unsupported)"),
}

func validateOptions() error {
	if *opts.yes && *opts.no {
		return errors.New("use of mutually exclusive -yes and -no")
	}
	if *opts.indent < 1 {
		return errors.New("-indent must be >= 1")
	}
	return nil
}
