// ac is a basic architectural calculator.
//
// The ac command supports the usual arithmetic operations over numbers
// and dimensions of the form
//
//    <n>' <n>"
//
// Any number can also include a fractional part separated by a space,
// such as "1 1/2". The fractional part must not include any spaces.
//
// Examples:
//
//    9' - 20"
//    = 7' 4"
//
//    8' 1 1/2" / 2
//    = 4' 3/4"
//
// ac uses arbitrary-precision rational numbers for all arithmetic.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/peterh/liner"
)

func main() {
	flag.Parse()
	if flag.NArg() > 0 {
		// Command-line mode
		if do(strings.Join(flag.Args(), " ")) != nil {
			os.Exit(1)
		}
		return
	}

	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	for {
		if s, err := line.Prompt("➤ "); err == nil {
			line.AppendHistory(s)
			do(s)
		} else {
			fmt.Printf("\n")
			if err == liner.ErrPromptAborted || err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
}

func do(s string) error {
	if s := strings.TrimSpace(s); len(s) == 0 {
		return nil
	}

	result, err := Parse(s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		if err, ok := err.(*SyntaxError); ok {
			fmt.Fprintf(os.Stderr, "\t%s\n\t%*s^\n", s, utf8.RuneCountInString(s[:err.pos]), "")
		}
		return err
	}
	fmt.Println(result)
	return nil
}
