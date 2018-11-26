package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	a := assert.New(t)

	exit = func(int) {}

	// Read in file to format
	srcData, err := ioutil.ReadFile("../test-files/fmtme.go")
	a.NoError(err)

	// Read in formatted file
	fmtData, err := ioutil.ReadFile("../test-files/nofmt.go")
	a.NoError(err)

	// copy the template file to a temp file
	tmp, err := ioutil.TempFile("", "*.go")
	a.NoError(err)

	tmpPath := tmp.Name()
	tmpDir := filepath.Dir(tmpPath)
	tmpFile := filepath.Base(tmpPath)

	_, err = tmp.Write(srcData)
	a.NoError(err)

	err = tmp.Close()
	a.NoError(err)

	setIn := func(newStdin *os.File) *os.File {
		stdin := os.Stdin
		os.Stdin = newStdin
		return stdin
	}
	restorIn := func(stdin *os.File) {
		os.Stdin = stdin
	}

	setOut := func(newStdout *os.File) *os.File {
		stdout := os.Stdout
		os.Stdout = newStdout
		return stdout
	}
	restorOut := func(stdout *os.File) {
		os.Stdout = stdout
	}

	setErr := func(newStderr *os.File) *os.File {
		stderr := os.Stderr
		os.Stderr = newStderr
		return stderr
	}
	restorErr := func(stderr *os.File) {
		os.Stderr = stderr
	}

	t.Run("stdin", func(t *testing.T) {
		a := assert.New(t)

		stdin, unfmted, err := os.Pipe()
		a.NoError(err)

		fmtted, stdout, err := os.Pipe()
		a.NoError(err)

		defer restorIn(setIn(stdin))
		defer restorOut(setOut(stdout))

		os.Args = []string{"nofmt"}

		// write the source data to stdin
		_, err = unfmted.Write(srcData)
		a.NoError(err)
		unfmted.Close() // close stdin to flush

		main()
		stdout.Close()

		// read formatted from standard out
		b := make([]byte, 1024)
		n, err := fmtted.Read(b)
		a.NoError(err)

		fmtted.Close()
		stdin.Close()

		a.Equal(string(b[:n]), string(fmtData))
	})

	t.Run("stdin error", func(t *testing.T) {
		a := assert.New(t)

		fmtted, stdout, err := os.Pipe()
		a.NoError(err)

		errData, stderr, err := os.Pipe()
		a.NoError(err)

		defer restorIn(setIn(nil))
		defer restorOut(setOut(stdout))
		defer restorErr(setErr(stderr))

		os.Args = []string{"nofmt"}

		main()
		stdout.Close()
		stderr.Close()

		b := make([]byte, 1024)
		n, _ := fmtted.Read(b)
		a.Equal(0, n)

		e := make([]byte, 1024)
		n, err = errData.Read(e)
		a.NoError(err)
		a.NotEqual(0, n)

		errData.Close()
		fmtted.Close()

		e = e[:n]
		a.Contains(string(e), "stdin:")
	})

	t.Run("stderr", func(t *testing.T) {
		a := assert.New(t)

		stdin, unfmted, err := os.Pipe()
		a.NoError(err)

		fmtted, stdout, err := os.Pipe()
		a.NoError(err)

		errData, stderr, err := os.Pipe()
		a.NoError(err)

		defer restorIn(setIn(stdin))
		defer restorOut(setOut(stdout))
		defer restorErr(setErr(stderr))

		os.Args = []string{"nofmt", "-F", "dd count=0"} // make some stderr without err

		// write the source data to stdin
		_, err = unfmted.Write(srcData)
		a.NoError(err)
		unfmted.Close() // close stdin to flush

		main()
		stdout.Close()
		stderr.Close()

		b := make([]byte, 1024)
		n, _ := fmtted.Read(b)
		a.Equal(0, n)

		e := make([]byte, 1024)
		n, err = errData.Read(e)
		a.NoError(err)
		a.NotEqual(0, n)

		errData.Close()
		fmtted.Close()
		stdin.Close()
	})

	t.Run("list", func(t *testing.T) {
		a := assert.New(t)

		listing, stdout, err := os.Pipe()
		a.NoError(err)

		defer restorOut(setOut(stdout))

		os.Args = []string{"nofmt", "-l", tmpDir}

		main()
		stdout.Close()

		s := strings.Builder{}
		// read listing from standard out
		for {
			b := make([]byte, 1024)
			n, err := listing.Read(b)
			a.NoError(err)
			s.Write(b[:n])
			if n < 1024 {
				break
			}
		}

		listing.Close()

		a.Contains(s.String(), tmpFile)

	})

	t.Run("diff", func(t *testing.T) {
		a := assert.New(t)

		diff, stdout, err := os.Pipe()
		a.NoError(err)

		defer restorOut(setOut(stdout))

		os.Args = []string{"nofmt", "-d", tmpPath}

		main()
		stdout.Close()

		b := make([]byte, 1024)
		n, err := diff.Read(b)
		a.NoError(err)
		b = b[:n]

		diff.Close()

		a.Contains(string(b), "--- "+tmpPath)
		a.Contains(string(b), "+++ "+tmpPath)
	})

	t.Run("diff fail", func(t *testing.T) {
		a := assert.New(t)

		diff, stdout, err := os.Pipe()
		a.NoError(err)

		defer restorOut(setOut(stdout))

		os.Args = []string{"nofmt", "-d", "-D", "foo/bar/no/diff/here"} // doesn't make diff output

		main()
		stdout.Close()

		b := make([]byte, 1024)
		n, _ := diff.Read(b)
		b = b[:n]

		diff.Close()

		a.NotContains(string(b), "--- "+tmpPath)
		a.NotContains(string(b), "+++ "+tmpPath)
	})

	t.Run("rewrite fail", func(t *testing.T) {
		a := assert.New(t)

		err := os.Chmod(tmpPath, 0444)
		a.NoError(err)

		os.Args = []string{"nofmt", "-w", tmpPath}

		main()

		err = os.Chmod(tmpPath, 0644)
		a.NoError(err)

		b, err := ioutil.ReadFile(tmpPath)
		a.NoError(err)

		a.Equal(string(srcData), string(b))
	})

	t.Run("rewrite", func(t *testing.T) {
		a := assert.New(t)

		os.Args = []string{"nofmt", "-w", tmpPath}

		main()

		b, err := ioutil.ReadFile(tmpPath)
		a.NoError(err)
		a.Equal(string(fmtData), string(b))
	})
}

func TestWalk(t *testing.T) {
	a := assert.New(t)

	ch := make(chan string, 128)
	go func() {
		walk(ch, []string{".", "", "/dev/null", "nofmt.go", "no/such/file/or/directory"})
		close(ch)
	}()

	sl := make([]string, 128)
	for s := range ch {
		sl = append(sl, s)
	}

	a.Contains(sl, "nofmt.go")
	a.Contains(sl, "nofmt_test.go")
	a.Contains(sl, "options.go")
	a.Contains(sl, "options_test.go")
	a.NotContains(sl, "Makefile")
	a.NotContains(sl, "..")
	a.NotContains(sl, ".")
}
