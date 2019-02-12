package parser

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatter(t *testing.T) {
	t.Run("Default Options, Missing File", func(t *testing.T) {
		f := NewFormatter("")
		a := assert.New(t)
		a.Equal(DefaultFmter, f.formatter)

		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		err := f.FormatFile("test-files/missing.go", stdout, stderr)
		assert.Error(t, err)
	})

	t.Run("Bad Fmter", func(t *testing.T) {
		f := NewFormatter("call-to-a-non-existant/fmt-program/should-fail-to-exec %f")
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		err := f.FormatFile("test-files/fmtme.go", stdout, stderr)
		assert.Error(t, err)
	})

	t.Run("Fmter Error", func(t *testing.T) {
		f := NewFormatter("cat -?") // make some stderr
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		err := f.FormatFile("test-files/fmtme.go", stdout, stderr)
		assert.Error(t, err)
		assert.NotEmpty(t, stderr)
	})

	t.Run("Fmter mismatch", func(t *testing.T) {
		f := NewFormatter("ls") // just data from formatter
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		err := f.FormatFile("test-files/fmtme.go", stdout, stderr)
		assert.Error(t, err)
	})

	t.Run("Fmt File", func(t *testing.T) {
		a := assert.New(t)

		f := NewFormatter("gofmt %f")
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		err := f.FormatFile("test-files/fmtme.go", stdout, stderr)
		a.NoError(err)

		expected, err := ioutil.ReadFile("test-files/nofmt.go")
		a.NoError(err)
		a.Equal(stdout.String(), string(expected))

		data, err := ioutil.ReadFile("test-files/fmtme.go")
		a.NoError(err)
		a.Equal(f.SourceData(), data)
	})

	t.Run("Fmt Stdin", func(t *testing.T) {
		a := assert.New(t)

		f := NewFormatter("gofmt %f")
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}

		dot, err := os.Open(".")
		a.NoError(err)

		err = f.FormatReader(dot, stdout, stderr)
		a.Error(err)
	})
}

func TestReadFile(t *testing.T) {
	fp, err := os.Open("test-files/fmtme.go")
	buf := bufio.NewReader(fp)

	blocks, err := readFile(buf)
	t.Log(err)
	for i, b := range blocks {
		t.Log(i, b)
	}

}

func TestProcessLine(t *testing.T) {

	// go:nofmt
	tests := []struct {
		name string
		in   codeState // Code, BlockComment, BackTick
		out  codeState // Code, BlockComment, BackTick, Fmt, NoFmt
		line string
	}{
		{name: "Normal Comment", in: Code, out: Code, line: "//build +nofmt"},
		{name: "Some Code", in: Code, out: Code, line: " 	foo := 1"},
		{name: "NoFmt - no space", in: Code, out: NoFmt, line: " 	//go:nofmt"},
		{name: "NoFmt - with space", in: Code, out: NoFmt, line: " 	// go:nofmt "},
		{name: "Fmt - no space", in: Code, out: Fmt, line: "//go:fmt"},
		{name: "Fmt - with space", in: Code, out: Fmt, line: " 	// go:fmt "},
		{name: "Normal comment", in: Code, out: Code, line: "  // not go:fmt"},
		{name: "Comment on a line", in: Code, out: Code, line: " foo = 3 // go:nofmt"},
		{name: "Comment on a line with slash", in: Code, out: Code, line: " foo = 3/4 // go:nofmt"},
		{name: "Block comment on a line", in: Code, out: Code, line: " foo := 1 /* set foo = 1 */"},
		{name: "Block comment short", in: Code, out: Code, line: "/**/"},
		{name: "Block comment start", in: Code, out: BlockComment, line: "  foo := 1 /* with tick `"},
		{name: "Block comment end", in: BlockComment, out: Code, line: "  with tick `*/"},
		{name: "BackTick on a line", in: Code, out: Code, line: " foo = `/* set foo = 1`"},
		{name: "BackTick start", in: Code, out: BackTick, line: "  foo = `1 /* with comment"},
		{name: "BackTick end", in: BackTick, out: Code, line: "  with comment /*`"},
		{name: "BackTick continue", in: BackTick, out: BackTick, line: "  /* comment in tick "},
		{name: "Quote", in: Code, out: Code, line: `baz := "normal quote"`},
		{name: "Quote in quote", in: Code, out: Code, line: `baz := "quote\" in quote\""`},
		{name: "Quote", in: Code, out: Code, line: `baz := "normal quote"`},
		{name: "Tick", in: Code, out: Code, line: `rune := 'x'`},
		{name: "Tick in tick", in: Code, out: Code, line: `rune := '\''`},
		{name: "Backslash in tick", in: Code, out: Code, line: `rune := '\\'`},
	}
	// go:fmt

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := parseLine(test.line, test.in)
			assert.Equal(t, test.out, out)
		})
	}

}

func TestFmtFail(t *testing.T) {
	f := Formatter{
		file:      "test-files/missing.go",
		formatter: "gofmt %f",
	}
	var errOut bytes.Buffer

	data, err := f.fmtFile(&errOut)
	assert.Error(t, err)
	assert.Empty(t, data)
}

func TestFmtFile(t *testing.T) {
	t.Run("File", func(t *testing.T) {
		f := Formatter{
			file:      "test-files/fmtme.go",
			formatter: "gofmt test-files/fmtme.go",
		}
		var errOut bytes.Buffer

		data, err := f.fmtFile(&errOut)
		assert.NoError(t, err)
		assert.Empty(t, errOut)
		assert.NotEmpty(t, data)
	})

	t.Run("Stdin", func(t *testing.T) {
		// Test with Standard In
		f := Formatter{
			file:      "",
			formatter: "gofmt test-files/fmtme.go",
			srcData:   *bytes.NewBufferString("package main\nfunc main(){}\n"),
		}
		var errOut bytes.Buffer

		data, err := f.fmtFile(&errOut)
		assert.NoError(t, err)
		assert.Empty(t, errOut)
		assert.NotEmpty(t, data)
	})
}
