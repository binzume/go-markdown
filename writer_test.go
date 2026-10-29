package markdown

import (
	"bytes"
	"testing"
)

func TestWriterInterface(t *testing.T) {
	var _ DocWriter = NewHTMLWriter(nil)
	var _ DocWriter = NewPlainWriter(nil)
}

type expectfun struct {
	fun      func(w DocWriter)
	expected string
}

func TestHtmlWriter(t *testing.T) {
	{
		var out bytes.Buffer
		writer := NewHTMLWriter(&out)

		expected := "Hello world!"
		writer.Write("Hello world!")
		writer.Close()
		actual := out.String()
		if actual != expected {
			t.Errorf("got %v\nwant %v", actual, expected)
		}
	}
}

func TestHTMLWriterHeadingOffset(t *testing.T) {
	var out bytes.Buffer
	writer := NewHTMLWriterWithHeadingOffset(&out, 1)

	n := writer.Heading(1, MarkupData{})
	writer.Write("Heading")
	writer.End(n)

	if got, want := out.String(), "<h2>Heading</h2>"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestHTMLWriterHeadingOffsetField(t *testing.T) {
	var out bytes.Buffer
	writer := NewHTMLWriter(&out)
	writer.HeadingOffset = 1

	n := writer.Heading(2, MarkupData{})
	writer.Write("Heading")
	writer.End(n)

	if got, want := out.String(), "<h3>Heading</h3>"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestPlainWriter(t *testing.T) {
	{
		var out bytes.Buffer
		writer := NewPlainWriter(&out)
		expected := "Hello world!"
		writer.Write("Hello world!")
		writer.Close()
		actual := out.String()
		if actual != expected {
			t.Errorf("got %v\nwant %v", actual, expected)
		}
	}
	{
		// TODO:
		tests := []expectfun{
			expectfun{func(w DocWriter) { w.Write("Hello") }, "Hello"},
			expectfun{func(w DocWriter) { w.WriteStyle("Hello", MarkupData{}) }, "Hello"},
			expectfun{func(w DocWriter) { w.Strike() }, ""},
			expectfun{func(w DocWriter) { w.Emphasis() }, ""},
			expectfun{func(w DocWriter) { w.Strong() }, ""},
			expectfun{func(w DocWriter) { w.Code() }, ""},
			expectfun{func(w DocWriter) { w.Block(MarkupData{Attributes: Attributes(".note")}) }, ""},
			expectfun{func(w DocWriter) { w.Link("http://example.com", "test", MarkupData{}) }, "http://example.com"},
			expectfun{func(w DocWriter) { w.Image("http://example.com/a.png", "", "test", MarkupData{}) }, "test(http://example.com/a.png)"},
			expectfun{func(w DocWriter) {
				n := w.Heading(1, MarkupData{})
				w.Write("test")
				w.End(n)
				w.Write("\n")
			}, "test\n"},
			expectfun{func(w DocWriter) { w.Paragraph() }, "\n"},
			expectfun{func(w DocWriter) { w.List(0) }, "\n"},
			expectfun{func(w DocWriter) { w.ListItem() }, ""},
			expectfun{func(w DocWriter) { w.Table() }, "\n"},
			expectfun{func(w DocWriter) { w.TableRow() }, "\n"},
			expectfun{func(w DocWriter) { w.TableCell(0) }, "\t"},
			expectfun{func(w DocWriter) { w.CodeBlock("golang", "test") }, "\n"},
			expectfun{func(w DocWriter) { w.Hr() }, ""},
		}

		for _, test := range tests {
			var out bytes.Buffer
			writer := NewPlainWriter(&out)
			test.fun(writer)
			writer.Close()
			actual := out.String()
			if actual != test.expected {
				t.Errorf("got %v\nwant %v", actual, test.expected)
			}
		}
	}
}
