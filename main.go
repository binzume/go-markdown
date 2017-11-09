package main

import (
	"bufio"
	"bytes"
	"os"
	"text/template"
)

const htmltemplate = `<html>
<head><link rel="stylesheet" type="text/css" href="theme/style.css" /></head>
<body><div class="gomd">{{markdown}}</div></body>
</html>`

func main() {

	funcMap := template.FuncMap{
		"markdown": MdRender,
	}
	t := template.Must(template.New("mdtest").Funcs(funcMap).Parse(htmltemplate))
	err := t.Execute(os.Stdout, map[string]interface{}{"TestValue": 3})
	if err != nil {
		panic(err)
	}
}

func MdRender() string {
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

	// out := os.Stdout
	var out bytes.Buffer

	scanner := bufio.NewScanner(fp)
	err = Convert(scanner, NewHTMLWriter(&out))
	if err != nil {
		panic(err)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return out.String()
}
