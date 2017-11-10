package main

import (
	"fmt"
	"html"
	"io"
)

// HTMLWriter : impl for DocWriter
type HTMLWriter struct {
	writer    io.Writer
	closetags []string
}

var DUMMY_DEPTH = 999999

func NewHTMLWriter(writer io.Writer) *HTMLWriter {
	return &HTMLWriter{writer, make([]string, 10)}
}

func (w *HTMLWriter) Heading(text string, level int) int {
	h := fmt.Sprint(level)
	io.WriteString(w.writer, "<h"+h+">"+html.EscapeString(text)+"</h"+h+">\n")
	return DUMMY_DEPTH
}

func (w *HTMLWriter) Paragraph() int {
	io.WriteString(w.writer, "<p>")
	w.closetags = append(w.closetags, "</p>\n")
	return len(w.closetags) - 1
}

func (w *HTMLWriter) Link(url string, title string, opt int) int {
	w.writer.Write([]byte("<a href='" + html.EscapeString(url) + "'>"))
	w.closetags = append(w.closetags, "</a>")
	return len(w.closetags) - 1
}

func (w *HTMLWriter) Image(url string, title string, opt int) int {
	w.writer.Write([]byte("<img src='" + html.EscapeString(url) + "' />"))
	return DUMMY_DEPTH
}

func (w *HTMLWriter) List(mode int) int {
	if mode == 0 {
		w.writer.Write([]byte("<ul>\n"))
		w.closetags = append(w.closetags, "</ul>\n")
	} else {
		w.writer.Write([]byte("<ol>\n"))
		w.closetags = append(w.closetags, "</ol>\n")
	}
	return len(w.closetags) - 1
}

func (w *HTMLWriter) ListItem() int {
	w.writer.Write([]byte("<li>"))
	w.closetags = append(w.closetags, "</li>\n")
	return len(w.closetags) - 1
}

func (w *HTMLWriter) Table() int {
	w.writer.Write([]byte("<table>\n"))
	w.closetags = append(w.closetags, "</table>\n")
	return len(w.closetags) - 1
}

func (w *HTMLWriter) TableRow() int {
	w.writer.Write([]byte("<tr>"))
	w.closetags = append(w.closetags, "</tr>\n")
	return len(w.closetags) - 1
}

func (w *HTMLWriter) TableCell(flags int) int {
	if flags == 1 {
		w.writer.Write([]byte("<th>"))
		w.closetags = append(w.closetags, "</th>")
	} else {
		w.writer.Write([]byte("<td>"))
		w.closetags = append(w.closetags, "</td>")
	}
	return len(w.closetags) - 1
}

func (w *HTMLWriter) Strike() int {
	w.writer.Write([]byte("<strike>"))
	w.closetags = append(w.closetags, "</strike>")
	return len(w.closetags) - 1
}

func (w *HTMLWriter) Strong() int {
	w.writer.Write([]byte("<strong>"))
	w.closetags = append(w.closetags, "</strong>")
	return len(w.closetags) - 1
}

func (w *HTMLWriter) Bold() int {
	w.writer.Write([]byte("<b>"))
	w.closetags = append(w.closetags, "</b>")
	return len(w.closetags) - 1
}

func (w *HTMLWriter) Italic() int {
	w.writer.Write([]byte("<i>"))
	w.closetags = append(w.closetags, "</i>")
	return len(w.closetags) - 1
}

func (w *HTMLWriter) Code() int {
	w.writer.Write([]byte("<code>"))
	w.closetags = append(w.closetags, "</code>")
	return len(w.closetags) - 1
}

func (w *HTMLWriter) CodeBlock(lang string, title string) int {
	w.writer.Write([]byte("<pre><code class='code_" + lang + "' title='" + html.EscapeString(title) + "'>"))
	w.closetags = append(w.closetags, "</code></pre>\n")
	return len(w.closetags) - 1
}

func (w *HTMLWriter) WriteStyle(text string, className string, color string, flags int) {
	attr := ""
	if className != "" {
		attr += " class='" + className + "'"
	}
	if color != "" {
		attr += " style='color:" + color + "'"
	}
	w.writer.Write([]byte("<span" + attr + ">"))
	w.Write(text)
	w.writer.Write([]byte("</span>"))
}

func (w *HTMLWriter) Write(text string) {
	w.writer.Write([]byte(html.EscapeString(text)))
}

func (w *HTMLWriter) End(lv int) {
	for len(w.closetags) > lv {
		io.WriteString(w.writer, w.closetags[len(w.closetags)-1])
		w.closetags = w.closetags[:len(w.closetags)-1]
	}
}

func (w *HTMLWriter) Close() {
	w.End(0)
}
