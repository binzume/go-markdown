package markdown

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

type kv struct {
	k string
	v string
}

func buildTag(tag string, end string, attrs ...kv) string {
	for _, kv := range attrs {
		if kv.v != "" {
			tag += " " + kv.k + "='" + html.EscapeString(kv.v) + "'"
		}
	}
	return tag + end
}

func (w *HTMLWriter) closeTag(t string) int {
	w.closetags = append(w.closetags, t)
	return len(w.closetags) - 1
}
func (w *HTMLWriter) simple(t string) int {
	io.WriteString(w.writer, "<"+t+">")
	return w.closeTag("</" + t + ">")
}

func (w *HTMLWriter) Heading(text string, level int) int {
	h := fmt.Sprint(level)
	io.WriteString(w.writer, "<h"+h+">"+html.EscapeString(text)+"</h"+h+">\n")
	return DUMMY_DEPTH
}

func (w *HTMLWriter) Paragraph() int {
	io.WriteString(w.writer, "<p>")
	return w.closeTag("</p>\n")
}

func (w *HTMLWriter) Link(url string, title string, opt int) int {
	io.WriteString(w.writer, buildTag("<a", ">", kv{"href", url}, kv{"title", title}))
	return w.closeTag("</a>")
}

func (w *HTMLWriter) Image(url string, title, alt string, opt int) int {
	io.WriteString(w.writer, buildTag("<img", "/>", kv{"src", url}, kv{"alt", alt}, kv{"title", title}))
	return DUMMY_DEPTH
}

func (w *HTMLWriter) Hr() int {
	io.WriteString(w.writer, "<hr/>")
	return DUMMY_DEPTH
}

func (w *HTMLWriter) List(mode int) int {
	if mode == 0 {
		w.writer.Write([]byte("<ul>\n"))
		return w.closeTag("</ul>\n")
	}
	w.writer.Write([]byte("<ol>\n"))
	return w.closeTag("</ol>\n")
}

func (w *HTMLWriter) ListItem() int {
	w.writer.Write([]byte("<li>"))
	return w.closeTag("</li>\n")
}

func (w *HTMLWriter) Table() int {
	w.writer.Write([]byte("<table>\n"))
	return w.closeTag("</table>\n")
}

func (w *HTMLWriter) TableRow() int {
	w.writer.Write([]byte("<tr>"))
	return w.closeTag("</tr>\n")
}

func (w *HTMLWriter) TableCell(flags int) int {
	style := []string{"", "text-align:left", "text-align:right", "text-align:center"}[flags&3]
	if flags&4 != 0 {
		io.WriteString(w.writer, buildTag("<th", ">", kv{"style", style}))
		return w.closeTag("</th>")
	}
	io.WriteString(w.writer, buildTag("<td", ">", kv{"style", style}))
	return w.closeTag("</td>")
}

func (w *HTMLWriter) CheckBox(checked bool) int {
	checkedStr := ""
	if checked {
		checkedStr = "checked"
	}
	io.WriteString(w.writer, buildTag("<input", "/>", kv{"type", "checkbox"}, kv{"checked", checkedStr}))
	return DUMMY_DEPTH
}

func (w *HTMLWriter) Strike() int {
	return w.simple("strike")
}

func (w *HTMLWriter) Emphasis() int {
	return w.simple("em")
}

func (w *HTMLWriter) Strong() int {
	return w.simple("strong")
}

func (w *HTMLWriter) Code() int {
	return w.simple("code")
}

func (w *HTMLWriter) QuoteBlock() int {
	return w.simple("blockquote")
}

func (w *HTMLWriter) CodeBlock(lang string, title string) int {
	if lang != "" {
		lang = "lang_" + lang
	}
	io.WriteString(w.writer, buildTag("<pre><code", ">", kv{"class", lang}, kv{"title", title}))
	return w.closeTag("</code></pre>\n")
}

func (w *HTMLWriter) WriteStyle(text string, className string, color string, flags int) {
	style := ""
	if color != "" {
		style += "color:" + color
	}
	io.WriteString(w.writer, buildTag("<span", ">", kv{"class", className}, kv{"style", style}))
	w.Write(text)
	w.writer.Write([]byte("</span>"))
}

func (w *HTMLWriter) Write(text string) {
	io.WriteString(w.writer, html.EscapeString(text))
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
