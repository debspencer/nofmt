package nofmt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var (
	// DefaultFmter is the default 'fmt' program The format
	// program will take either a file name or standard in and
	// format it into Golang syntax Arguments are separated by
	// spaces.  If %f appears in the argument string will be
	// replaced by the filename.  If the file is stdandard input
	// no file will provided.
	DefaultFmter = "gofmt"
)

// block will contains slices of the file to be formatted.  A file can
// be separated into alternating blocks of formatted and unformatted
// code.
type block struct {
	formatted bool
	lines     []string
}

// Formatter contains information about the file being formatted
type Formatter struct {
	file      string       // name of file, blank for standard in
	formatter string       // formatter with arguments
	original  []*block     // original file with formatted and unformatted blocks (note: formatted blocks are to be formatted)
	processed []*block     // post processed file with formatted and unformatted blocks
	srcData   bytes.Buffer // original source data of file
}

// file is a path to file that needs fmting.  If file is empty, stdin is assumed

// New returns a Formatter object with the default fmter
func New() *Formatter {
	return &Formatter{
		formatter: DefaultFmter,
	}
}

// NewFormatter returns a Formatter object
// Formatter program options.  Replace %f with filename if present or append if not.
// examples: "gofmt %f", "gofmt", "/home/go/bin/goimports", "myfmttool -f %f -pretty"
// If stdin is used in %f will br replaced with a enpty string
func NewFormatter(formatter string) *Formatter {
	if len(formatter) == 0 {
		return New()
	}

	return &Formatter{
		formatter: formatter,
	}
}

// FormatFile will write fmted output from file to the out io.Writer
// If there are syntax errors in the file and it can not be formatted, then error text will be written to errOut
// An error can be returned without any data being written to errOur
// Format will scan file for pramga codes // go:nofmt and // go:fmt
func (f *Formatter) FormatFile(file string, out io.Writer, errOut io.Writer) error {

	f.file = file

	// open source file
	fp, err := os.Open(f.file)
	if err != nil {
		return err
	}
	defer fp.Close()
	return f.FormatReader(fp, out, errOut)
}

// FormatReader will write fmted output from reader to to the out io.Writer
// If there are syntax errors in the file and it can not be formatted, then error text will be written to errOut
// An error can be returned without any data being written to errOur
// Format will scan file for pramga codes // go:nofmt and // go:fmt
func (f *Formatter) FormatReader(in io.Reader, out io.Writer, errOut io.Writer) error {
	_, err := io.Copy(&f.srcData, in)
	if err != nil {
		return err
	}

	// Read the source file and determine nofmt blocks
	// all blocks will be unformtted, but marked formatted or unformatted blocks
	orig := bufio.NewReader(bytes.NewBuffer(f.srcData.Bytes()))
	f.original, _ = readFile(orig)

	// run "fmt" on the file
	formatted, err := f.fmtFile(errOut)
	if err != nil {
		return err
	}

	// prococess the fmt file into formated and unformatted blocks
	// all blocks will be formtted, but marked formatted or unformatted blocks
	proc := bufio.NewReader(formatted)
	f.processed, _ = readFile(proc) // there is no way this can fail on a buffer

	// We should have the same number of blocks before and after
	if len(f.original) != len(f.processed) {
		return fmt.Errorf("block mismatch: %d != %d", len(f.original), len(f.processed))
	}

	// Write out the fmtted data.
	// The formatted blocks from the fmter
	// The unformatted blocks from the original
	for i := range f.processed {
		var lines []string
		if f.processed[i].formatted {
			lines = f.processed[i].lines
		} else {
			lines = f.original[i].lines
		}
		for l := range lines {
			out.Write([]byte(lines[l]))
		}
	}
	return nil
}

// SourceData returns the original source data
func (f *Formatter) SourceData() []byte {
	return f.srcData.Bytes()
}

type codeState int

const (
	Code codeState = iota
	NoFmt
	Fmt
	BlockComment
	BackTick
)

// readFile will read a go source file and return a set of blocks (collection of lines)
// each block will alternate between formatted and unformatted code.
func readFile(buf *bufio.Reader) ([]*block, error) {
	curState := Code

	curBlock := &block{
		formatted: true,
		lines:     make([]string, 0, 1024),
	}
	blocks := make([]*block, 0, 16)
	blocks = append(blocks, curBlock)

	var err error

	// Read each line
	for {
		var line string
		line, err = buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}

		// parse the code line.  After parsing the line the code can be in one of several states:
		// Code:         Normal code state - add code to current block
		// NoFmt:        Found a // go:nofmt marker - switch to a unformatted block
		// Fmt:          Found a // go:fmt marker - switch to a formatted block
		// BlockComment  Inside a block commend /* */ - add code to current block
		// BackTick      Inside a back tick string `` - add code to current block
		newState := parseLine(line, curState)
		if newState != curState {
			switch newState {
			case NoFmt:
				curState = Code
				if curBlock.formatted {
					// encoutered a go:nofmt block
					// add the control line to the current formatted block and create a new one
					curBlock.lines = append(curBlock.lines, line)
					curBlock = &block{
						formatted: false,
						lines:     make([]string, 0, 128),
					}
					blocks = append(blocks, curBlock)
					// continue since we already added the line
					continue
				}
			case Fmt:
				curState = Code
				if !curBlock.formatted {
					// encoutered a go:fmt block
					// create a new block
					curBlock = &block{
						formatted: true,
						lines:     make([]string, 0, 1024),
					}
					blocks = append(blocks, curBlock)
				}
			default:
				curState = newState
			}
		}
		curBlock.lines = append(curBlock.lines, line)

	}
	return blocks, err
}

type lineState int64

const (
	Other            lineState = iota // before finding anything on the line
	Indent                            // before finding anything on the line
	BeginSlash                        // Found a slash (first on the line)
	FoundSlash                        // Found a slash (not the first slash)
	FoundStar                         // Found a * in an Block comment
	LineComment                       // inside a line comment //
	BeginLineComment                  // found a Line comment at the being of the line
	InBlockComment                    // is a block comment /*
	InBackTick                        // Inside a `string`
	InQuote                           // Inside a "quote"
	InTick                            // Instde a 'tick'
	FoundQuote                        // Inside a "quote", found a quote, peek back for \
	FoundTick                         // Inside a 'tick', found a tick, peek back for \
	EndComment                        // End of a Block comment
	EndBackTick                       // End of a Back tick block
)

type runeState map[rune]lineState

type nextState struct {
	found    map[rune]lineState
	notFound lineState
}

var (
	// as we process a line, depending on the character we encounter, we need to transition to a new state.
	// found will be a map of the current rune and the next state to transition to.
	// notFound will be a state to transtion to if a character is not found
	stateTable = map[lineState]nextState{
		Other: nextState{
			found: runeState{
				'/':  FoundSlash,
				'`':  InBackTick,
				'"':  InQuote,
				'\'': InTick,
			},
		},
		Indent: nextState{
			found: runeState{
				' ':  Indent,
				'\t': Indent,
				'/':  BeginSlash,
			},
		},
		BeginSlash: nextState{
			found: runeState{
				'/': BeginLineComment, // this is an end state
				'*': InBlockComment,
			},
		},
		FoundSlash: nextState{
			found: runeState{
				'/': LineComment, // this is an end state
				'*': InBlockComment,
			},
		},
		FoundStar: nextState{
			found: runeState{
				'/': EndComment,
			},
			notFound: InBlockComment,
		},
		InBlockComment: nextState{
			found: runeState{
				'*': FoundStar,
			},
			notFound: InBlockComment,
		},
		InBackTick: nextState{
			found: runeState{
				'`': EndBackTick,
			},
			notFound: InBackTick,
		},
		InQuote: nextState{
			found: runeState{
				'"': FoundQuote, // look for \
			},
			notFound: InQuote,
		},
		InTick: nextState{
			found: runeState{
				'\'': FoundTick, // look for \
			},
			notFound: InTick,
		},
	}
)

// parseLine will evaluate a line to see if fmt needs to be enabled or disabled
// curState will be the codeState from the previous line
func parseLine(line string, curState codeState) codeState {
	// We start assuming we are ready for an Indent, or code, unless the previous line
	// found us in a Block Comment or Back Tick quote.
	st := Indent
	switch curState {
	case BlockComment:
		st = InBlockComment
	case BackTick:
		st = InBackTick
	}

	table := stateTable[st]

	for i, r := range line {
		// peer into the state table for the current rune being processed
		// if not found, take the notFound entry (which defaults to Other)
		newState, ok := table.found[r]
		if !ok {
			newState = table.notFound
		}

		switch newState {
		case Other:
			// Other is just normal code. Set the return state
			curState = Code
		case BeginLineComment:
			// Found a // comment at the begining of the line
			// if it is a go:fmt or go:nofmt pragma, signal immediately we have found a marker
			comment := strings.TrimSpace(line[i+1:])
			switch comment {
			case "go:fmt":
				return Fmt
			case "go:nofmt":
				return NoFmt
			}
			return Code
		case LineComment:
			return Code
		case InBlockComment:
			curState = BlockComment
		case InBackTick:
			curState = BackTick
		case EndComment, EndBackTick:
			curState = Code
			st = Other
		case FoundQuote, FoundTick:
			// We hit an end " or or end ' mark.  Before we return to code, we need to make sure the
			// previous character was not an escape
			n := 0
			for j := i - 1; j >= 0; j-- {
				if line[j] != '\\' {
				}
				n++
			}
			if n&1 == 0 {
				st = Other
			}
		}
		if st != newState {
			st = newState
			table = stateTable[st]
		}
	}

	return curState
}

// run the fmter of the source file and capture the output
func (f *Formatter) fmtFile(errOut io.Writer) (*bytes.Buffer, error) {

	// add the file to the formatter
	formatter := strings.TrimSpace(f.formatter)
	if strings.Contains(f.formatter, "%f") {
		formatter = strings.Replace(formatter, "%f", f.file, -1)
	} else {
		formatter = formatter + " " + f.file
	}
	formatter = strings.TrimSpace(formatter)

	// split the command line in two to get the command
	cmdArgs := strings.SplitN(formatter, " ", 2)

	// split the args in an array
	var args []string
	if len(cmdArgs) > 1 {
		args = strings.Split(cmdArgs[1], " ")
	}

	cmd := exec.Command(cmdArgs[0], args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	var errChan chan error
	if f.file == "" {
		stdin, err := cmd.StdinPipe()
		errChan = make(chan error)
		go func() {
			if err == nil {
				_, err = io.Copy(stdin, bytes.NewBuffer(f.srcData.Bytes()))
				stdin.Close()
			}
			errChan <- err
		}()
	}

	err := cmd.Run()
	if stderr.Len() > 0 {
		errOut.Write(stderr.Bytes())
		errStr := ""
		if err != nil {
			errStr = fmt.Sprintf(" (%s)", err.Error())
		}
		return nil, fmt.Errorf("%s: returned error%s", formatter, errStr)
	}
	if err != nil {
		return nil, err
	}
	if errChan != nil {
		err = <-errChan
	}
	return &stdout, err
}
