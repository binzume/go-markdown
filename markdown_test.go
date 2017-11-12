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
		expect{`**hello**`, `<p><strong>hello</strong></p>`},
		expect{`*hello*`, `<p><em>hello</em></p>`},
		expect{`~~**hello**~~`, `<p><strike><strong>hello</strong></strike></p>`},
		expect{"`this is code.`", `<p><code>this is code.</code></p>`},
		expect{"``this is `code`.``", "<p><code>this is `code`.</code></p>"},
		expect{"url: http://www.example.com/?hello", "<p>url: <a href='http://www.example.com/?hello'>http://www.example.com/?hello</a></p>"},
		expect{`[link](test.png)`, "<p><a href='test.png'>link</a></p>"},
		expect{`[link](test.png "test")`, "<p><a href='test.png' title='test'>link</a></p>"},
		expect{`![img](test.png)`, "<p><img src='test.png' alt='img'/></p>"},
		expect{`![img](test.png "test")`, "<p><img src='test.png' alt='img' title='test'/></p>"},
		expect{`[![img](test.png)](test)`, "<p><a href='test'><img src='test.png' alt='img'/></a></p>"},
		expect{`[![img](test.png) ![img](test.png)](test)`, "<p><a href='test'><img src='test.png' alt='img'/> <img src='test.png' alt='img'/></a></p>"},
		// block
		expect{"# hello", `<h1>hello</h1>`},
		expect{"## hello", `<h2>hello</h2>`},
		expect{"----------", "<hr/>"},
		expect{"> quote", "<blockquote>quote\n</blockquote>"},
		expect{"|a|b|\n|-|-|\n|1|2|\n", "<table>\n<tr><th>a</th><th>b</th></tr>\n<tr><td>1</td><td>2</td></tr>\n</table>"},
		expect{"- item1\n- item2\n", "<ul>\n<li>item1</li>\n<li>item2</li>\n</ul>"},
		expect{"- [ ] hoge", "<ul>\n<li><input type='checkbox'/>hoge</li>\n</ul>"},
		expect{"[dummy]: # (dummy ref)", ""},
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

func TestCodeBlock(t *testing.T) {
	in := strings.NewReader("```go\n// test\nfunc main() {\nfmt.Print(\"hello!\")\n}\n```")
	expected := strings.Replace(
		`<pre><code class='lang_go'><span class='code_comment'>// test</span>
		<span class='code_key'>func</span> <span class='code_ident'>main</span>() {
		<span class='code_ident'>fmt</span>.<span class='code_ident'>Print</span>(<span class='code_str'>&#34;hello!&#34;</span>)
		}
		</code></pre>`, "\t", "", -1)

	var out bytes.Buffer
	writer := NewHTMLWriter(&out)
	err := Convert(bufio.NewScanner(in), writer)
	if err != nil {
		t.Errorf("error %v", err)
	}
	writer.Close()

	if strings.TrimSpace(out.String()) != expected {
		t.Errorf("got '%v'\nwant '%v'", out.String(), expected)
	}
}
