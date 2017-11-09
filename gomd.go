package main

import (
	"bufio"
	"regexp"
	"strings"
)

type Matcher interface {
	Prefix() string
	TryMatch(line string) (int, []string)
	Render(params []string, scanner *bufio.Scanner, writer DocWriter)
}

type SimpleInlineMatcher struct {
	Start      string
	End        string
	RenderFunc func(content string, writer DocWriter, matcher *SimpleInlineMatcher)
}

func (m *SimpleInlineMatcher) Prefix() string {
	return m.Start
}

func (m *SimpleInlineMatcher) TryMatch(line string) (int, []string) {
	p := strings.Index(line, m.End)
	if p < 0 {
		return -1, nil
	}
	return p + len(m.End), []string{line[:p]}
}

func (m *SimpleInlineMatcher) Render(params []string, scanner *bufio.Scanner, writer DocWriter) {
	m.RenderFunc(params[0], writer, m)
}

type RegexMatcher struct {
	Re         *regexp.Regexp
	RenderFunc func(matches []string, writer DocWriter, scanner *bufio.Scanner, matcher *RegexMatcher)
}

func (m *RegexMatcher) Prefix() string {
	return ""
}

func (m *RegexMatcher) TryMatch(text string) (int, []string) {
	match := m.Re.FindStringSubmatch(text)
	if len(match) < 1 {
		return -1, nil
	}
	return len(match[0]), match
}

func (m *RegexMatcher) Render(params []string, scanner *bufio.Scanner, writer DocWriter) {
	m.RenderFunc(params, writer, scanner, m)
}

type LinkInlineMatcher struct{}

func (m *LinkInlineMatcher) Prefix() string {
	return "["
}

func (m *LinkInlineMatcher) TryMatch(line string) (int, []string) {
	p1 := strings.Index(line, "](")
	if p1 < 0 {
		return -1, nil
	}
	p2 := strings.Index(line, ")")
	if p2 < 0 || p2 < p1 {
		return -1, nil
	}
	return p2 + 1, []string{line[:p1], line[p1+2 : p2]}
}

func (m *LinkInlineMatcher) Render(params []string, scanner *bufio.Scanner, writer DocWriter) {
	n := writer.Link(params[1], 0)
	writer.Write(params[0])
	writer.End(n)
}

func strike(text string, writer DocWriter, markup *SimpleInlineMatcher) {
	a := writer.Strike()
	writer.Write(text)
	writer.End(a)
}

func bold(text string, writer DocWriter, markup *SimpleInlineMatcher) {
	a := writer.Bold()
	writer.Write(text)
	writer.End(a)
}

func icode(text string, writer DocWriter, markup *SimpleInlineMatcher) {
	a := writer.Code()
	writer.Write(text)
	writer.End(a)
}

func autolink(params []string, writer DocWriter, scanner *bufio.Scanner, markup *RegexMatcher) {
	n := writer.Link(params[0], 0)
	writer.Write(params[0])
	writer.End(n)
}

var inlineElems = []Matcher{
	&SimpleInlineMatcher{"~", "~", strike},
	&SimpleInlineMatcher{"*", "*", bold},
	&SimpleInlineMatcher{"`", "`", icode},
	&SimpleInlineMatcher{"__", "__", bold},
	&LinkInlineMatcher{},
	&RegexMatcher{regexp.MustCompile(`^https?:[^\s\"\']+`), autolink},
}

func inline(text string, writer DocWriter) {
	for pos := 0; pos < len(text); pos++ {
		for _, markup := range inlineElems {
			// TODO more fast.
			if strings.HasPrefix(text[pos:], markup.Prefix()) {
				l, params := markup.TryMatch(text[(pos + len(markup.Prefix())):])
				if l > 0 {
					writer.Write(text[:pos])
					markup.Render(params, nil, writer)
					l += len(markup.Prefix())
					text = text[(pos + l):]
					pos = 0
				}
			}
		}
	}
	writer.Write(text)
}

func list(params []string, writer DocWriter, scanner *bufio.Scanner, markup *RegexMatcher) {
	nt := writer.List()
	for scanner.Scan() {
		ni := writer.ListItem()
		inline(params[2], writer)
		writer.End(ni)

		text := scanner.Text()
		params = markup.Re.FindStringSubmatch(text)
		if len(params) < 1 {
			break
		}
	}
	writer.End(nt)
}

func code(params []string, writer DocWriter, scanner *bufio.Scanner, markup *RegexMatcher) {
	lang := params[1]
	n := writer.CodeBlock(lang)
	for scanner.Scan() {
		text := scanner.Text()
		m := markup.Re.FindStringSubmatch(text)
		if len(m) > 0 {
			break
		}
		tokenizer := NewTokenizer(lang)
		tokenizer.Code(text)
		for typ, s := tokenizer.Read(); typ != CODE_EOF; typ, s = tokenizer.Read() {
			switch typ {
			case CODE_Keyword:
				writer.WriteStyle(s, "code_key", "", 0)
			case CODE_Number:
				writer.WriteStyle(s, "code_num", "", 0)
			case CODE_String:
				writer.WriteStyle(s, "code_str", "", 0)
			case CODE_Comment:
				writer.WriteStyle(s, "code_comment", "", 0)
			default:
				writer.Write(s)
			}
		}
		writer.Write("\n")
	}
	writer.End(n)
}

func table(params []string, writer DocWriter, scanner *bufio.Scanner, markup *RegexMatcher) {
	nt := writer.Table()
	for scanner.Scan() {
		text := params[1]
		nr := writer.TableRow()
		for _, s := range strings.Split(text, "|") {
			nc := writer.TableCell()
			inline(s, writer)
			writer.End(nc)
		}
		writer.End(nr)

		text = scanner.Text()
		params = markup.Re.FindStringSubmatch(text)
		if len(params) < 1 {
			break
		}
	}
	writer.End(nt)
}

func heading(params []string, writer DocWriter, scanner *bufio.Scanner, markup *RegexMatcher) {
	writer.Heading(params[2], len(params[1]))
}

func comment(params []string, writer DocWriter, scanner *bufio.Scanner, markup *RegexMatcher) {
}

func pluginBlock(params []string, writer DocWriter, scanner *bufio.Scanner, markup *RegexMatcher) {
	// TODO
	n := writer.CodeBlock("")
	for scanner.Scan() {
		text := scanner.Text()
		if text == "}" {
			break
		}
	}
	writer.End(n)
}

var blockElem = []Matcher{
	&RegexMatcher{regexp.MustCompile(`^(#{1,4})\s*(.*)`), heading},
	&RegexMatcher{regexp.MustCompile(`^//.*`), comment},
	&RegexMatcher{regexp.MustCompile("^```\\s*(\\w*)$"), code},
	&RegexMatcher{regexp.MustCompile(`^\|(.+)\|$`), table},
	&RegexMatcher{regexp.MustCompile(`^(\s*)-\s?(.+)$`), list},
	&RegexMatcher{regexp.MustCompile(`^&(\w+){?$`), pluginBlock},
}

// ToHTML convert md to html
func Convert(scanner *bufio.Scanner, writer DocWriter) error {
	para := 0
	for scanner.Scan() {
		text := scanner.Text()
		for _, matcher := range blockElem {
			l, params := matcher.TryMatch(text)
			if l > 0 {
				if para != 0 {
					writer.End(para)
					para = 0
				}
				writer.Write("\n")
				matcher.Render(params, scanner, writer)
				text = ""
				break
			}
		}
		if text == "" {
			if para != 0 {
				writer.End(para)
				para = 0
			}
			continue
		}
		if para == 0 {
			para = writer.Paragraph()
		}
		inline(text, writer)
	}
	writer.Close()
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
