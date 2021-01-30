package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yevhenshymotiuk/asana-to-org/tasks"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}

func run() error {
	fp := os.Args[len(os.Args)-1]
	ofp := flag.String(
		"o",
		fmt.Sprintf("%s.org", strings.TrimSuffix(fp, filepath.Ext(fp))),
		"Output file",
	)
	includeComleted := flag.Bool("wc", false, "Include completed tasks")
	flag.Parse()

	os.Setenv("INCLUDE_COMPLETED", strconv.FormatBool(*includeComleted))

	f, err := os.Open(fp)
	defer f.Close()
	if err != nil {
		return err
	}

	d, err := tasks.NewData(io.Reader(f))
	if err != nil {
		return err
	}

	data := *d
	sections := tasks.NewSections(data.Data)

	of, err := os.Create(*ofp)
	defer of.Close()
	if err != nil {
		return err
	}
	of.WriteString(sections.String())

	return nil
}
