package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"text/template"

	"github.com/binzume/go-markdown"
)

const htmltemplate = `<html>
<head><link rel="stylesheet" type="text/css" href="theme/style.css" /></head>
<body><div class="gomd">{{.input | markdown}}</div></body>
</html>`

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

	funcMap := template.FuncMap{
		"markdown": MdRender,
	}
	t := template.Must(template.New("mdtest").Funcs(funcMap).Parse(htmltemplate))
	err = t.Execute(os.Stdout, map[string]interface{}{"TestValue": 3, "input": fp})
	if err != nil {
		panic(err)
	}
}

func MdRender(in io.Reader) string {
	// out := os.Stdout
	var out bytes.Buffer

	scanner := bufio.NewScanner(in)
	writer := markdown.NewHTMLWriter(&out)
	err := markdown.Convert(scanner, writer)
	if err != nil {
		panic(err)
	}
	writer.Close()

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return out.String()
}
