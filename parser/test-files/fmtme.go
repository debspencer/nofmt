package main

import "fmt"

func main() {
fmt.Println("hello \"world\"")
fmt.Println(`hello world\
		// go:nofmt
`)

// go:fmt

// this is a comment
/* this is a comment
	// go:nofmt

	*/
// go:nofmt
type foo struct {
	a  int
	bb string
}

// go:fmt
// go:fmt
type baz struct {
	a  int
	bb string
}
// go:nofmt

fmt.Println(&foo{})
fmt.Println(&baz{})
}
