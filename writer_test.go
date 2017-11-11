package markdown

import (
	"bytes"
	"testing"
)

func TestHtmlWriter(t *testing.T) {
	{
		var out bytes.Buffer
		writer := NewHTMLWriter(&out)
		expected := "Hello world!"
		var a interface{} = writer
		if _, ok := a.(DocWriter); !ok {
			t.Errorf("not DocWriter")
		}

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
		var a interface{} = writer
		if _, ok := a.(DocWriter); !ok {
			t.Errorf("not DocWriter")
		}

		writer.Write("Hello world!")
		writer.Close()
		actual := out.String()
		if actual != expected {
			t.Errorf("got %v\nwant %v", actual, expected)
		}

	}
}
