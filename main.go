package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/echlebek/grill/grill"
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
		err = f.Close()
	}()

	r := grill.NewReader(f)
	var (
		t     grill.Test
		tests []grill.Test
	)

	for err == nil {
		err = r.Read(&t)

		t.Filepath = path
		tests = append(tests, t)
	}
	if err != io.EOF && err != nil {
		return nil, err
	}

	return &grill.TestSuite{Tests: tests, Name: path}, nil
}

func Main(a []string, stdout, stderr io.Writer) int {
	if err := flags.Parse(a); err != nil {
		stderr.Write([]byte(err.Error()))
		return 2
	}
	args := flags.Args()
	if *opts.version {
		fmt.Fprintln(stderr, grillVersion)
		return 0
	}
	if len(args) == 0 {
		fmt.Fprint(stderr, "Usage: grill [OPTIONS] TESTS...\n")
		return 2
	}

	context, err := grill.DefaultTestContext("bash", stdout, stderr)
	if err != nil {
		log.Println(err)
		return 1
	}

	rc := 0

	for _, a := range args {
		suite, err := readTestSuite(a)
		if err != nil {
			rc = 1
			log.Println(err)
			continue
		}

		for i := range suite.Tests {
			err := suite.Tests[i].Run(context)
			if err != nil {
				rc = 1
				log.Println(err)
			}
		}
		if suite.Failed() {
			rc = 1
			if err := suite.WriteErr(); err != nil {
				log.Println(err)
			}
		}
		if err := suite.WriteReport(stderr); err != nil {
			log.Println(err)
		}
	}

	return rc
}
