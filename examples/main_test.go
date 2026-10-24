package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestNoteDirective(t *testing.T) {
	got := MdRender(strings.NewReader(":::note custom\n**message**\n:::"))
	want := "\n<div class='note-custom'>\n<p><strong>message</strong></p>\n</div>\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestExamples(t *testing.T) {
	input, err := os.ReadFile("sample.md")
	if err != nil {
		t.Fatal(err)
	}

	got := MdRender(bytes.NewReader(input))
	if !strings.Contains(got, "<div class='note-info'>") {
		t.Fatalf("note directive was not rendered")
	}

	output, err := os.Create("sample.html")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	fmt.Fprint(output, `<html><head><link rel="stylesheet" type="text/css" href="theme/style.css" /></head>
		<body><div class="gomd">`)
	fmt.Fprint(output, got)
	fmt.Fprint(output, `</div></body></html>`)
}
