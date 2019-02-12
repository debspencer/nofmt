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

## Usage

```
usage: nofmt [-d|-w|-l] [-D <diffprog>] [-e] [-F <fmter>] [file|dir ...]
  -D string
        diff program to use
  -F string
        specify formatter 'program args' (filename will be appended unless %f is used) (default "gofmt %f")
  -d    only show differences
  -e    pass -e to formatter program
  -l    list all files whose formatting differs from nofmt's
  -w    write back to file(s) instead of stdout
  ```

#### `-D string`
When using `-d` diff option, specify an alternate diff program to use
to generate diffs.  Default diff program is `$PATH/diff -u`.  To pass
options to diff program enclose program name in quotes.  By default
files passed to the diff program are appended to the end of the prgram
provided.  To specify file order use `%f1` and `%f2` for placeholders of
file names.

Examples:
`nofmt -d -D /opt/bsd/diff -u foo.go`
`nofmt -f -D 'sdiff -w132' foo.go`
`nofmt -f -D 'sdiff -w132 %f1 %f2' foo.go`

#### `-F string`

Specify the formatter to use.  `nofmt` uses a formatter to do the
formatting for regions that are not between `// go:nofmt` pragmas.
Default formatter is `gofmt`, but can plug any format program such as
`goimports`.  To specify a file to the formatter use `%f` otherwise
the filename will be appened the command.

Examples:
`nofmt -F gofmt foo.go`
`nofmt -F goimports foo.go`
`nofmt -F 'myformater -f %f' foo.go`

#### `-d`

Show differences between the current file(s) and formatted version.
Use `-D` to specify a diff program other than `diff`.

#### `-e`

Pass `-e` option to formatter.  Both `gofmt` and `goimports` use `-e`
to report more than just 10 errors.

#### `-l`

List all files whose formatting differs from that of `nofmt`.

#### `-w`

Write formatting changes back to original source file and not to
stdout.  Must specify source file.

#### `file|dir ...`

One or more files or directories can be specified (can mix).  If a
directory is specified, `nofmt` will walk the directory and apply
options to each file with a `.go` extension.  If no files or
directories are given, the `nofmt` will operate on stdin.`

## License
This project is provide AS-IS.  Please see [../LICENSE](LICENSE) file.


## Contributing

Please feel free to submit issues, fork the repository and send pull requests!  Feedback is always welcome.

When submitting an issue, please include an example that demonstrates the issue.
