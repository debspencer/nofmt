package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/debspencer/diff"
	"github.com/debspencer/nofmt"
)

func main() {
	opt := getOptions(os.Args)

	files := make(chan string, 16)
	if len(opt.files) == 0 {
		opt.write = false
		go func() {
			files <- ""
			close(files)
		}()
	} else {
		go func() {
			walk(files, opt.files)
			close(files)
		}()
	}

	for file := range files {
		fmter := nofmt.NewFormatter(opt.formatter)
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}

		var err error
		if file == "" {
			err = fmter.FormatReader(os.Stdin, stdout, stderr)
		} else {
			err = fmter.FormatFile(file, stdout, stderr)
		}
		if err != nil {
			if file == "" {
				file = "stdin"
			}
			fmt.Fprintf(os.Stderr, "%s: %s\n", file, err)
			continue
		}

		if opt.diff {
			if file == "" {
				file = "<stdin>"
			}
			b1 := diff.Buffer{Data: fmter.SourceData(), Filename: file + ".orig"}
			b2 := diff.Buffer{Data: stdout.Bytes(), Filename: file}

			diffData, err := diff.DiffBuffer(b1, b2)

			if err != nil {
				fmt.Fprintf(os.Stderr, "diff failed: %s\n", err)
			} else {
				fmt.Print(string(diffData))
			}
			continue
		}

		if opt.write {
			mode := os.FileMode(0644)
			st, err := os.Stat(file)
			if err == nil {
				mode = st.Mode()
			}
			err = ioutil.WriteFile(file, stdout.Bytes(), mode)
			if err != nil {
				fmt.Fprintf(os.Stderr, "rewrite %s: %s\n", file, err)
			}
			continue
		}
		if opt.list {
			if bytes.Compare(fmter.SourceData(), stdout.Bytes()) != 0 {
				fmt.Println(file)
			}
			continue
		}
		fmt.Print(stdout.String())
	}
}

func walk(ch chan string, files []string) {
	for _, file := range files {
		if len(file) == 0 {
			continue
		}
		fi, err := os.Stat(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		mode := fi.Mode()
		if mode.IsRegular() {
			ch <- file
			continue
		}
		if mode.IsDir() {
			filepath.Walk(file, func(path string, info os.FileInfo, err error) error {
				if info != nil && info.Mode().IsRegular() && info.Size() > 0 && filepath.Ext(path) == ".go" {
					ch <- path
				}
				return nil
			})
			continue
		}
		fmt.Fprintf(os.Stderr, "%s unsupport mode %s\n", fi.Name(), mode)
	}
}
