package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/echlebek/grill/grill"
)

func main() {
	flag.Parse()
	if err := Main(os.Args, os.Stdout, os.Stderr); err != nil {
		log.Fatal(err)
	}
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
		tests = append(tests, t)
	}
	if err != io.EOF && err != nil {
		return nil, err
	}

	return &grill.TestSuite{Tests: tests, Name: path}, nil
}

func Main(args []string, stdout, stderr io.Writer) error {
	args = flag.Args()
	if len(args) == 0 {
		fmt.Fprint(stdout, "Usage: grill [OPTIONS] TESTS...\n")
		os.Exit(2)
	}
	context, err := grill.DefaultTestContext(".", "bash", stdout, stderr)
	if err != nil {
		return err
	}

	for _, a := range args {
		suite, err := readTestSuite(a)
		if err != nil {
			log.Println(err)
		}

		for i := range suite.Tests {
			err := suite.Tests[i].Run(context)
			if err != nil {
				log.Println(err)
			}
		}
		if err := suite.WriteErr(); err != nil {
			log.Println(err)
		}
		if err := suite.WriteReport(stderr); err != nil {
			log.Println(err)
		}
	}

	return nil
}
