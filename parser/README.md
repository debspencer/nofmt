# Nofmt - Format your code the way you want it

Nofmt is a drop in replacement for existing formatting tools such as `gofmt` or `goimports`.

Nofmt will format your code using a formatter tool, `gofmt` by
default, however it will not format code blocks which are bracketed by
a `// go:nofmt` and `// go:fmt` pragma.

For example, `gofmt` would take the following code

```go
package main

import "fmt"

func main() {
        var s                string
        var longVariableName string

        fmt.Prihtln(s, longVariableName)
}
```

And produce the following:
```go
package main

import "fmt"

func main() {
        var s string
        var longVariableName string

        fmt.Prihtln(s, longVariableName)
}
```

With nofmt's pragma's you can keep the code formatted like you want it.

```go
package main

import "fmt"

func main() {
        // go:nofmt
        var s                string
        var longVariableName string
        // go:fmt

        fmt.Prihtln(s, longVariableName)
}
```

This can be very useful for certain sparsly populataed data stuctures, where alignment can aide readability.

See [nofmt/README.md](nofmt/README.md) file for info on using `nofmt` program.


# nofmt
`import "github.com/debspencer/nofmt"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)
* [Subdirectories](#pkg-subdirectories)

## <a name="pkg-overview">Overview</a>



## <a name="pkg-index">Index</a>
* [Constants](#pkg-constants)
* [Variables](#pkg-variables)
* [type Formatter](#Formatter)
  * [func New() *Formatter](#New)
  * [func NewFormatter(formatter string) *Formatter](#NewFormatter)
  * [func (f *Formatter) FormatFile(file string, out io.Writer, errOut io.Writer) error](#Formatter.FormatFile)
  * [func (f *Formatter) FormatReader(in io.Reader, out io.Writer, errOut io.Writer) error](#Formatter.FormatReader)
  * [func (f *Formatter) SourceData() []byte](#Formatter.SourceData)


#### <a name="pkg-files">Package files</a>
[nofmt.go](/src/github.com/debspencer/nofmt/nofmt.go) 

## <a name="pkg-constants">Constants</a>
``` go
const (
    Code codeState = iota
    NoFmt
    Fmt
    BlockComment
    BackTick
)
```
``` go
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
```

## <a name="pkg-variables">Variables</a>
``` go
var (
    // DefaultFmter is the default 'fmt' program The format
    // program will take either a file name or standard in and
    // format it into Golang syntax Arguments are separated by
    // spaces.  If %f appears in the argument string will be
    // replaced by the filename.  If the file is stdandard input
    // no file will provided.
    DefaultFmter = "gofmt"
)
```

## <a name="Formatter">type</a> [Formatter](/src/target/nofmt.go?s=716:1121#L32)
``` go
type Formatter struct {
    // contains filtered or unexported fields
}

```
Formatter contains information about the file being formatted

### <a name="New">func</a> [New](/src/target/nofmt.go?s=1262:1283#L43)
``` go
func New() *Formatter
```
New returns a Formatter object with the default fmter

### <a name="NewFormatter">func</a> [NewFormatter](/src/target/nofmt.go?s=1616:1662#L53)
``` go
func NewFormatter(formatter string) *Formatter
```
NewFormatter returns a Formatter object
Formatter program options.  Replace %f with filename if present or append if not.
examples: "gofmt %f", "gofmt", "/home/go/bin/goimports", "myfmttool -f %f -pretty"
If stdin is used in %f will br replaced with a enpty string

### <a name="Formatter.FormatFile">func</a> (\*Formatter) [FormatFile](/src/target/nofmt.go?s=2079:2161#L67)
``` go
func (f *Formatter) FormatFile(file string, out io.Writer, errOut io.Writer) error
```
FormatFile will write fmted output from file to the out io.Writer
If there are syntax errors in the file and it can not be formatted, then error text will be written to errOut
An error can be returned without any data being written to errOur
Format will scan file for pramga codes // go:nofmt and // go:fmt

### <a name="Formatter.FormatReader">func</a> (\*Formatter) [FormatReader](/src/target/nofmt.go?s=2650:2735#L84)
``` go
func (f *Formatter) FormatReader(in io.Reader, out io.Writer, errOut io.Writer) error
```
FormatReader will write fmted output from reader to to the out io.Writer
If there are syntax errors in the file and it can not be formatted, then error text will be written to errOut
An error can be returned without any data being written to errOur
Format will scan file for pramga codes // go:nofmt and // go:fmt

### <a name="Formatter.SourceData">func</a> (\*Formatter) [SourceData](/src/target/nofmt.go?s=3976:4015#L129)
``` go
func (f *Formatter) SourceData() []byte
```
SourceData returns the original source data

## License
This project is provide AS-IS.  Please see [LICENSE](LICENSE) file.

## Contributing

Please feel free to submit issues, fork the repository and send pull requests!  Feedback is always welcome.

When submitting an issue, please include an example that demonstrates the issue.
