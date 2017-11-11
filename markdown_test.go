package markdown

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

type expect struct {
	input    string
	expected string
}

func TestConvert(t *testing.T) {

	tests := []expect{
		// inline
		expect{`hello`, `<p>hello</p>`},
		expect{`~~hello~~`, `<p><strike>hello</strike></p>`},
		expect{`*hello*`, `<p><em>hello</em></p>`},
		expect{`**hello**`, `<p><strong>hello</strong></p>`},
		expect{"`this is code.`", `<p><code>this is code.</code></p>`},
		expect{"``this is `code`.``", "<p><code>this is `code`.</code></p>"},
		expect{"url: http://www.example.com/", "<p>url: <a href='http://www.example.com/'>http://www.example.com/</a></p>"},
		// block
		expect{"# hello", `<h1>hello</h1>`},
		expect{"## hello", `<h2>hello</h2>`},
		expect{"> quote", "<blockquote>quote\n</blockquote>"},
		expect{"- list", "<ul>\n<li>list</li>\n</ul>"},
	}

	for _, test := range tests {
		in := strings.NewReader(test.input)
		var out bytes.Buffer
		writer := NewHTMLWriter(&out)
		err := Convert(bufio.NewScanner(in), writer)
		if err != nil {
			t.Errorf("error %v", err)
		}
		writer.Close()

		if strings.TrimSpace(out.String()) != test.expected {
			t.Errorf("got '%v'\nwant '%v'", out.String(), test.expected)
		}

	}
}
