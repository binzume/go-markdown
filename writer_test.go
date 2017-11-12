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
			expectfun{func(w DocWriter) { w.WriteStyle("Hello", "", "", 0) }, "Hello"},
			expectfun{func(w DocWriter) { w.Strike() }, ""},
			expectfun{func(w DocWriter) { w.Emphasis() }, ""},
			expectfun{func(w DocWriter) { w.Strong() }, ""},
			expectfun{func(w DocWriter) { w.Code() }, ""},
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
