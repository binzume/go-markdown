package main

import (
	"bytes"
	"io"
	"os"
	"text/template"

	"github.com/binzume/go-markdown"
	"github.com/binzume/go-markdown/directives"
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
	var out bytes.Buffer

	md := markdown.NewMarkdown()
	directives.RegisterNote(md)

	writer := markdown.NewHTMLWriter(&out)
	err := md.Convert(in, writer)
	if err != nil {
		panic(err)
	}
	writer.Close()

	return out.String()
}
