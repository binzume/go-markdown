package markdown

import (
	"bufio"
	"bytes"
	"io"
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
		expect{"hello\nworld", "<p>hello\nworld</p>"},
		expect{"hello  \nworld", "<p>hello<br>\nworld</p>"},
		expect{"hello\n\nworld", "<p>hello</p>\n<p>world</p>"},
		expect{`~~hello~~`, `<p><strike>hello</strike></p>`},
		expect{`**hello**`, `<p><strong>hello</strong></p>`},
		expect{`*hello*`, `<p><em>hello</em></p>`},
		expect{`~~**hello**~~`, `<p><strike><strong>hello</strong></strike></p>`},
		expect{"`this is code.`", `<p><code>this is code.</code></p>`},
		expect{"``this is `code`.``", "<p><code>this is `code`.</code></p>"},
		expect{`\*escaped*`, `<p>*escaped*</p>`},
		expect{`aaa ** bbb`, `<p>aaa ** bbb</p>`},
		expect{"url: http://www.example.com/?hello", "<p>url: <a href='http://www.example.com/?hello'>http://www.example.com/?hello</a></p>"},
		expect{`:unknown[body]{argument}`, "<p><span>body</span></p>"},
		expect{`prefix :unknown[**body**]{argument}`, "<p>prefix <span><strong>body</strong></span></p>"},
		expect{`[link](test.png)`, "<p><a href='test.png'>link</a></p>"},
		expect{`[link](test.png "test")`, "<p><a href='test.png' title='test'>link</a></p>"},
		expect{`[link](test.png){.className}`, "<p><a href='test.png' class='className'>link</a></p>"},
		expect{`[link](test.png){#linkId .className}`, "<p><a href='test.png' class='className' id='linkId'>link</a></p>"},
		expect{`![img](test.png)`, "<p><img src='test.png' alt='img'/></p>"},
		expect{`![img](test.png "test")`, "<p><img src='test.png' alt='img' title='test'/></p>"},
		expect{`![img](test.png){.thumbnail}`, "<p><img src='test.png' alt='img' class='thumbnail'/></p>"},
		expect{`![img](test.png){#imageId .thumbnail}`, "<p><img src='test.png' alt='img' class='thumbnail' id='imageId'/></p>"},
		expect{`[![img](test.png)](test)`, "<p><a href='test'><img src='test.png' alt='img'/></a></p>"},
		expect{`[![img](test.png) ![img](test.png)](test)`, "<p><a href='test'><img src='test.png' alt='img'/> <img src='test.png' alt='img'/></a></p>"},

		// block
		expect{"# hello", `<h1>hello</h1>`},
		expect{"## hello", `<h2>hello</h2>`},
		expect{"## Hedding {#elemId}", `<h2 id='elemId'>Hedding</h2>`},
		expect{"## Hedding {#elemId .title}", `<h2 class='title' id='elemId'>Hedding</h2>`},
		expect{`# **strong** *emphasis* [link](target)`, `<h1><strong>strong</strong> <em>emphasis</em> <a href='target'>link</a></h1>`},
		expect{"----------", "<hr/>"},
		expect{"> quote\n> aaa", "<blockquote>quote\naaa\n</blockquote>"},
		expect{"|a|b|\n|-|-|\n|1|2|\n", "<table>\n<tr><th>a</th><th>b</th></tr>\n<tr><td>1</td><td>2</td></tr>\n</table>"},
		expect{"- item1\n- item2\n", "<ul>\n<li>item1</li>\n<li>item2</li>\n</ul>"},
		expect{"- aaaa\n  bbbb\n- cccc", "<ul>\n<li>aaaa\nbbbb</li>\n<li>cccc</li>\n</ul>"},
		expect{"- aaaa  \n  bbbb\n- cccc", "<ul>\n<li>aaaa<br>\nbbbb</li>\n<li>cccc</li>\n</ul>"},
		expect{"1. item1\n2. item2\n", "<ol>\n<li>item1</li>\n<li>item2</li>\n</ol>"},
		expect{"- [ ] hoge", "<ul>\n<li><input type='checkbox'/>hoge</li>\n</ul>"},
		expect{"- [x] fuga", "<ul>\n<li><input type='checkbox' checked='checked'/>fuga</li>\n</ul>"},
		expect{"[dummy]: # (dummy ref)", ""},

		// code
		expect{"```go\n// test\nfunc main() {\nfmt.Print(\"hello!\")\n}\n```",
			strings.Replace(
				`<pre><code class='lang_go'><span class='code_comment'>// test</span>
				<span class='code_key'>func</span> <span class='code_ident'>main</span>() {
				<span class='code_ident'>fmt</span>.<span class='code_ident'>Print</span>(<span class='code_str'>&#34;hello!&#34;</span>)
				}
				</code></pre>`, "\t", "", -1)},
		expect{"``` rb\n# sample\ndef main()\nputs \"hello!\"\nend\n```",
			strings.Replace(
				`<pre><code class='lang_rb'><span class='code_comment'># sample</span>
				<span class='code_key'>def</span> <span class='code_ident'>main</span>()
				<span class='code_ident'>puts</span> <span class='code_str'>&#34;hello!&#34;</span>
				<span class='code_key'>end</span>
				</code></pre>`, "\t", "", -1)},
	}

	for _, test := range tests {
		in := strings.NewReader(test.input)
		var out bytes.Buffer
		writer := NewHTMLWriter(&out)
		err := Convert(in, writer)
		if err != nil {
			t.Errorf("error %v", err)
		}
		writer.Close()

		if strings.TrimSpace(out.String()) != test.expected {
			t.Errorf("got '%v'\nwant '%v'", out.String(), test.expected)
		}

	}
}

func TestDirectives(t *testing.T) {
	md := NewMarkdown()
	var body string
	md.RegisterDirective("note", func(d *DirectiveContext) error {
		content, err := io.ReadAll(d.Content)
		body = string(content)
		if d.Name != "note" || d.Argument != "warning compact" {
			t.Errorf("unexpected directive context: %#v", d)
		}
		return err
	})

	var out bytes.Buffer
	writer := NewHTMLWriter(&out)
	err := md.Convert(strings.NewReader(":::note warning compact\n**hello**\n:::"), writer)
	if err != nil {
		t.Fatal(err)
	}
	if body != "**hello**\n" {
		t.Errorf("got body %q", body)
	}
}

func TestInlineDirectives(t *testing.T) {
	md := NewMarkdown()
	md.RegisterDirective("badge", func(d *DirectiveContext) error {
		if !d.Inline || d.Name != "badge" || d.Argument != "warning" {
			t.Fatalf("unexpected directive context: %#v", d)
		}
		body, err := io.ReadAll(d.Content)
		if err != nil {
			return err
		}
		n := d.Writer.Span(MarkupData{Attributes: Attributes(".badge-warning")})
		d.Writer.Write(string(body))
		d.Writer.End(n)
		return nil
	})

	var out bytes.Buffer
	writer := NewHTMLWriter(&out)
	if err := md.Convert(strings.NewReader(":badge[body]{warning}"), writer); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	want := "<p><span class='badge-warning'>body</span></p>"
	if got := strings.TrimSpace(out.String()); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestMarkupDataIDs(t *testing.T) {
	var out bytes.Buffer
	writer := NewHTMLWriter(&out)

	n := writer.Span(MarkupData{Attributes: Attributes("#span-id")})
	writer.End(n)
	n = writer.Block(MarkupData{Attributes: Attributes("#block-id")})
	writer.End(n)
	writer.Close()

	want := "<span id='span-id'></span><div id='block-id'>\n</div>\n"
	if got := out.String(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAttributes(t *testing.T) {
	id, className := Attributes("#elemId .title .wide").GetIdAndClassName()
	if id != "elemId" || className != "title wide" {
		t.Fatalf("got id=%q, className=%q", id, className)
	}

	id, className = Attributes("unknown=value").GetIdAndClassName()
	if id != "" || className != "" {
		t.Fatalf("invalid attributes returned id=%q, className=%q", id, className)
	}

	id, className = Attributes("{#elemId .title}").GetIdAndClassName()
	if id != "elemId" || className != "title" {
		t.Fatalf("braced attributes returned id=%q, className=%q", id, className)
	}
}

func TestReferenceLinks(t *testing.T) {
	input := `[docs]: https://example.com/docs "Documentation"

[full][DOCS] [docs][] [docs] ![logo][docs]`
	var out bytes.Buffer
	writer := NewHTMLWriter(&out)
	if err := Convert(strings.NewReader(input), writer); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	want := `<p><a href='https://example.com/docs' title='Documentation'>full</a> <a href='https://example.com/docs' title='Documentation'>docs</a> <a href='https://example.com/docs' title='Documentation'>docs</a> <img src='https://example.com/docs' alt='logo' title='Documentation'/></p>`
	if got := strings.TrimSpace(out.String()); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestUnknownDirectiveRendersAsDiv(t *testing.T) {
	input := ":::unknown\n**hello**\n:::"
	var out bytes.Buffer
	writer := NewHTMLWriter(&out)
	if err := Convert(strings.NewReader(input), writer); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	want := `<div>
<p><strong>hello</strong></p>
</div>`
	if got := strings.TrimSpace(out.String()); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestNestedDirectivesUseLongerFences(t *testing.T) {
	md := NewMarkdown()
	var body string
	md.RegisterDirective("outer", func(d *DirectiveContext) error {
		content, err := io.ReadAll(d.Content)
		body = string(content)
		return err
	})

	var out bytes.Buffer
	input := "::::outer\n:::inner\ninside\n:::\nafter\n::::"
	writer := NewHTMLWriter(&out)
	if err := md.Convert(strings.NewReader(input), writer); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, ":::inner\ninside\n:::") {
		t.Errorf("nested directive body was not preserved: %q", body)
	}
}

func TestUnclosedDirective(t *testing.T) {
	md := NewMarkdown()
	md.RegisterDirective("note", func(d *DirectiveContext) error { return nil })
	var out bytes.Buffer
	err := md.Convert(strings.NewReader(":::note\nbody"), NewHTMLWriter(&out))
	if err == nil || !strings.Contains(err.Error(), "unclosed directive") {
		t.Fatalf("expected unclosed directive error, got %v", err)
	}
}

func TestDelimiterReaderStopsAtDelimiter(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("one\n```not-end\n```\nafter"))
	state := &SyntaxContext{Scanner: scanner}
	reader := &DelimiterReader{
		scanner:   state,
		delimiter: []byte("```"),
	}

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "one\n```not-end\n" {
		t.Errorf("got %q", got)
	}
	if rest, err := io.ReadAll(reader); err != nil || len(rest) != 0 {
		t.Errorf("reader continued after delimiter: %q, %v", rest, err)
	}
}
