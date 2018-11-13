package main

import (
	"flag"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testFlagError int
)

func testErrorHandler(n int) {
	testFlagError = n
}

func TestOptions(t *testing.T) {
	flagErrorHandling = flag.ContinueOnError
	flagErrorHandler = testErrorHandler

	// usage: goimports [-d|-w|-l] [-e] [-F <fmter>] [file|dir ...]
	tests := []struct {
		flags string
		opt   options
		error bool
	}{
		{flags: "", opt: options{formatter: "gofmt %f"}},
		{flags: "-d", opt: options{formatter: "gofmt %f", diff: true}},
		{flags: "-w", opt: options{formatter: "gofmt %f", write: true}, error: true},
		{flags: "-l", opt: options{formatter: "gofmt %f", list: true, files: []string{"."}}},
		{flags: "-d -l file", opt: options{formatter: "gofmt %f", diff: true, list: true, files: []string{"file"}}, error: true},
		{flags: "-l -w file", opt: options{formatter: "gofmt %f", list: true, write: true, files: []string{"file"}}, error: true},
		{flags: "-d -w file", opt: options{formatter: "gofmt %f", diff: true, write: true, files: []string{"file"}}, error: true},
		{flags: "-d -D diff_-u a b", opt: options{formatter: "gofmt %f", diff: true, differ: "diff -u", files: []string{"a", "b"}}},
		{flags: "-w -F myfmt file", opt: options{formatter: "myfmt", write: true, files: []string{"file"}}},
		{flags: "-e", opt: options{formatter: "gofmt -e %f", errors: true}},
		{flags: "-w -e -F myfmt file", opt: options{formatter: "myfmt -e", errors: true, write: true, files: []string{"file"}}},
	}
	for _, test := range tests {
		testFlagError = 0
		t.Run(test.flags, func(t *testing.T) {
			if test.opt.files == nil {
				test.opt.files = []string{}
			}
			prog := strings.TrimSpace("prog " + test.flags)

			opts := strings.Fields(prog)
			for i := range opts {
				opts[i] = strings.Replace(opts[i], "_", " ", -1)
			}
			test.opt.args = opts

			opt := getOptions(opts)
			opt.f = nil

			a := assert.New(t)
			a.Equal(test.opt, *opt)
			a.Equal(test.error, (testFlagError == 2))
		})
	}
}
