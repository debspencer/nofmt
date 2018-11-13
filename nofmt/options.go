package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/debspencer/diff"
)

var (
	flagErrorHandling = flag.ExitOnError
	flagErrorHandler  = os.Exit
)

type options struct {
	args      []string
	f         *flag.FlagSet
	diff      bool
	differ    string
	errors    bool
	files     []string
	formatter string
	write     bool
	list      bool
}

func getOptions(args []string) *options {
	f := flag.NewFlagSet(args[0], flagErrorHandling)

	o := &options{
		f:    f,
		args: args,
	}
	f.Usage = o.usage
	f.BoolVar(&o.diff, "d", false, "only show differences")
	f.StringVar(&o.differ, "D", "", "diff program to use")
	f.BoolVar(&o.errors, "e", false, "pass -e to formatter program")
	f.StringVar(&o.formatter, "F", "gofmt %f", "specify formatter 'program args' (filename will be appended unless %f is used)")
	f.BoolVar(&o.list, "l", false, "list all files whose formatting differs from nofmt's")
	f.BoolVar(&o.write, "w", false, "write back to file(s) instead of stdout")
	f.Parse(args[1:])
	o.files = f.Args()

	if o.write && len(o.files) == 0 {
		fmt.Fprintln(os.Stderr, "Can not rewrite <stdin>")
		o.usage()
	}

	if countBools(o.diff, o.list, o.write) > 1 {
		o.usage()
	}

	if o.list && len(o.files) == 0 {
		o.files = []string{"."}
	}

	if o.errors {
		if strings.Contains(o.formatter, " ") {
			o.formatter = strings.Replace(o.formatter, " ", " -e ", 1)
		} else {
			o.formatter += " -e"
		}
	}

	if o.diff && len(o.differ) > 0 {
		s := strings.Fields(o.differ)
		diff.DiffProgram = s[0]

		if len(s) > 1 {
			diff.DiffProgramArgs = strings.Join(s[1:], " ")
		} else {
			diff.DiffProgramArgs = ""
		}
	}

	return o
}

func (o *options) usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [-d|-w|-l] [-D <diffprog>] [-e] [-F <fmter>] [-n] [file|dir ...]\n", filepath.Base(o.args[0]))
	o.f.PrintDefaults()
	flagErrorHandler(2)
}

func countBools(bools ...bool) int {
	n := 0
	for _, b := range bools {
		if b {
			n++
		}
	}
	return n
}
