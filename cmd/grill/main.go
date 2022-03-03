package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/echlebek/grill/internal/grill"
)

const grillVersion = "dev"

func main() {
	os.Exit(Main(os.Args[1:], os.Stdout, os.Stderr))
}

func readTestSuite(path string) (ts *grill.TestSuite, err error) {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("couldn't read test file: %s", err)
		return nil, err
	}

	defer func() {
		if fErr := f.Close(); fErr != nil {
			// Don't clobber previous error
			if err == nil {
				err = fErr
			} else {
				log.Println(fErr)
			}
		}
	}()

	r := grill.NewReader(f)
	var (
		t     grill.Test
		tests []grill.Test
	)

	for err == nil {
		err = r.Read(&t)
		tests = append(tests, t)
	}
	if err != io.EOF && err != nil {
		return nil, err
	}

	return &grill.TestSuite{Tests: tests, Name: path}, nil
}

func Main(a []string, stdout, stderr io.Writer) int {
	if err := flags.Parse(a); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 2
	}

	if *opts.version {
		fmt.Fprintln(stderr, grillVersion)
		return 0
	}

	args := flags.Args()
	if len(args) == 0 {
		fmt.Fprint(stderr, "Usage: grill [OPTIONS] TESTS...\n")
		return 2
	}

	context, err := grill.DefaultTestContext(*opts.shell, *opts.preserveEnv)
	if err != nil {
		log.Println(err)
		return 1
	}

	defer func() {
		if *opts.keepTmpdir {
			if _, err := fmt.Fprintf(stdout, "# Kept temporary directory: %s\n", context.WorkDir); err != nil {
				log.Println(err)
			}
		} else {
			context.Cleanup()
		}
	}()

	var (
		rc     int
		suites []*grill.TestSuite
	)

	for _, a := range args {
		suite, err := readTestSuite(a)
		if err != nil {
			log.Println(err)
			return 1
		}

		suites = append(suites, suite)

		if err := suite.Run(context); err != nil {
			log.Println(err)
			return 1
		}

		if *opts.verbose {
			_, err = fmt.Fprintf(stdout, "%s: %s\n", suite.Name, suite.Status())
		} else {
			_, err = fmt.Fprint(stdout, suite.StatusGlyph())
		}
		if err != nil {
			log.Println(err)
			return 1
		}

		if suite.Failed() {
			rc = 1
			if err := suite.WriteErr(); err != nil {
				log.Println(err)
				return 1
			}
		} else {
			if err := suite.RemoveErr(); err != nil {
				log.Println(err)
				return 1
			}
		}
	}

	if !*opts.verbose {
		if _, err := fmt.Fprint(stdout, "\n"); err != nil {
			log.Println(err)
			return 1
		}
	}

	if err := grill.WriteReport(stdout, suites, *opts.quiet); err != nil {
		log.Println(err)
		return 1
	}

	return rc
}
