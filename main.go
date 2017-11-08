package main

import (
	"bufio"
	"os"
)

func main() {
	var fp *os.File
	var err error

	if len(os.Args) < 2 {
		fp = os.Stdin
	} else {
		fp, err = os.Open(os.Args[1])
		if err != nil {
			panic(err)
		}
		defer fp.Close()
	}

	scanner := bufio.NewScanner(fp)
	err = Convert(scanner, NewHTMLWriter(os.Stdout))
	if err != nil {
		panic(err)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
