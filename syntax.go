package markdown

import (
	"regexp"
	"strings"
)

type Syntax interface {
	Prefix() string
	TryParse(line string) (int, []string)
	Render(params []string, s *SyntaxContext)
}

type SimpleInlineSyntax struct {
	Start      string
	End        string
	RenderFunc func(content string, s *SyntaxContext, syntax *SimpleInlineSyntax)
}

func (m *SimpleInlineSyntax) Prefix() string { return m.Start }

func (m *SimpleInlineSyntax) TryParse(line string) (int, []string) {
	line = line[len(m.Start):]
	p := strings.Index(line, m.End)
	if p <= 0 {
		return -1, nil
	}
	return p + len(m.End) + len(m.Start), []string{line[:p]}
}

func (m *SimpleInlineSyntax) Render(params []string, s *SyntaxContext) {
	m.RenderFunc(params[0], s, m)
}

type RegexSyntax struct {
	PrefixStr  string
	Re         *regexp.Regexp
	RenderFunc func(matches []string, s *SyntaxContext, syntax *RegexSyntax)
}

func (m *RegexSyntax) Prefix() string { return m.PrefixStr }

func (m *RegexSyntax) TryParse(text string) (int, []string) {
	match := m.Re.FindStringSubmatch(text)
	if len(match) < 1 {
		return -1, nil
	}
	return len(match[0]), match
}

func (m *RegexSyntax) Render(params []string, s *SyntaxContext) {
	m.RenderFunc(params, s, m)
}

// HeadingSyntax matches ATX headings without invoking the regexp engine.
type HeadingSyntax struct{}

func (m *HeadingSyntax) Prefix() string { return "#" }

func (m *HeadingSyntax) TryParse(line string) (int, []string) {
	level := 0
	for level < len(line) && level < 4 && line[level] == '#' {
		level++
	}
	if level == 0 {
		return -1, nil
	}

	content := strings.TrimLeft(line[level:], " \t\r\f")
	return len(line), []string{line[:level], content}
}

func (m *HeadingSyntax) Render(params []string, s *SyntaxContext) {
	renderHeading(params[1], len(params[0]), s)
}

type LinkInlineSyntax struct{ Start string }

func (m *LinkInlineSyntax) Prefix() string { return m.Start }

func (m *LinkInlineSyntax) TryParse(line string) (int, []string) {
	p1 := strings.Index(line, "](")
	if p1 < 0 {
		return -1, nil
	}

	pos := len(m.Start)
retry:
	if strings.Index(line[pos:p1], "![") >= 0 {
		if p := strings.Index(line[p1+1:], "]("); p >= 0 {
			pos = p1
			p1 += p + 1
			goto retry
		}
	}
	p2 := strings.Index(line[p1:], ")") + p1
	if p2 < 0 {
		return -1, nil
	}

	end := p2 + 1
	attributes := ""
	if strings.HasPrefix(line[end:], "{") {
		if attrEnd := strings.IndexByte(line[end+1:], '}'); attrEnd >= 0 {
			attrEnd += end + 1
			if _, ok := Attributes(line[end+1 : attrEnd]).markupData(); ok {
				attributes = line[end+1 : attrEnd]
				end = attrEnd + 1
			}
		}
	}

	return end, []string{line[len(m.Start):p1], line[p1+2 : p2], attributes}
}

func (m *LinkInlineSyntax) Render(params []string, md *SyntaxContext) {
	url := strings.Split(params[1], " ")
	var title string
	if len(url) > 1 {
		title = strings.Trim(url[1], "\" ")
	}
	data := MarkupData{}
	if len(params) > 2 {
		data, _ = Attributes(params[2]).markupData()
	}
	if m.Start == "![" {
		n := md.Image(url[0], title, params[0], data)
		md.End(n)
		return
	}
	n := md.Link(url[0], title, data)
	md.inline(params[0])
	md.End(n)
}

type ReferenceLinkInlineSyntax struct {
	Start string
	Image bool
}

func (m *ReferenceLinkInlineSyntax) Prefix() string { return m.Start }

func (m *ReferenceLinkInlineSyntax) TryParse(line string) (int, []string) {
	if !strings.HasPrefix(line, m.Start) {
		return -1, nil
	}

	firstEnd := strings.IndexByte(line[len(m.Start):], ']')
	if firstEnd < 0 {
		return -1, nil
	}
	firstEnd += len(m.Start)
	text := line[len(m.Start):firstEnd]

	// An inline link is handled by LinkInlineSyntax.
	if firstEnd+1 < len(line) && line[firstEnd+1] == '(' {
		return -1, nil
	}

	end := firstEnd + 1
	label := text
	if end < len(line) && line[end] == '[' {
		refEnd := strings.IndexByte(line[end+1:], ']')
		if refEnd < 0 {
			return -1, nil
		}
		refEnd += end + 1
		if refEnd > end+1 {
			label = line[end+1 : refEnd]
		}
		end = refEnd + 1
	} else if m.Image {
		// Shortcut reference images are not valid without a label.
		return -1, nil
	}

	return end, []string{text, label, line[:end]}
}

func (m *ReferenceLinkInlineSyntax) Render(params []string, s *SyntaxContext) {
	var ok bool
	if m.Image {
		ok = s.WriteReferenceImage(params[0], params[1])
	} else {
		ok = s.WriteReferenceLink(params[0], params[1])
	}
	if !ok {
		s.Write(params[2])
	}
}

func strike(text string, md *SyntaxContext, syntax *SimpleInlineSyntax) {
	n := md.Strike()
	md.inline(text)
	md.End(n)
}

func strong(text string, md *SyntaxContext, syntax *SimpleInlineSyntax) {
	n := md.Strong()
	md.inline(text)
	md.End(n)
}

func emphasis(text string, md *SyntaxContext, syntax *SimpleInlineSyntax) {
	n := md.Emphasis()
	md.inline(text)
	md.End(n)
}

func icode(text string, md *SyntaxContext, syntax *SimpleInlineSyntax) {
	n := md.Code()
	md.Write(text)
	md.End(n)
}

func autolink(params []string, md *SyntaxContext, syntax *RegexSyntax) {
	n := md.Link(params[0], "", MarkupData{})
	md.Write(params[0])
	md.End(n)
}

func heading(params []string, md *SyntaxContext, syntax *RegexSyntax) {
	renderHeading(params[2], len(params[1]), md)
}

func renderHeading(text string, level int, s *SyntaxContext) {
	text, data := parseHeadingAttributes(text)
	n := s.Heading(level, data)
	s.inline(text)
	s.End(n)
	s.Write("\n")
}

// parseHeadingAttributes removes a trailing attribute block from a heading.
// Currently IDs and CSS classes are supported using the usual Markdown
// attribute syntax, for example: "Heading {#heading-id .section}".
func parseHeadingAttributes(text string) (string, MarkupData) {
	trimmed := strings.TrimRight(text, " \t\r\f")
	if !strings.HasSuffix(trimmed, "}") {
		return text, MarkupData{}
	}

	open := strings.LastIndexByte(trimmed, '{')
	if open <= 0 || !isWhitespace(trimmed[open-1]) {
		return text, MarkupData{}
	}

	data, ok := Attributes(trimmed[open+1 : len(trimmed)-1]).markupData()
	if !ok {
		return text, MarkupData{}
	}
	return strings.TrimRight(trimmed[:open], " \t\r\f"), data
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\f'
}

func hr(params []string, md *SyntaxContext, syntax *RegexSyntax) { md.Hr() }

func dummy(params []string, md *SyntaxContext, syntax *RegexSyntax) {}

func list(params []string, s *SyntaxContext, syntax *RegexSyntax) {
	var mode int
	switch params[2] {
	case "*", "-", "+":
		mode = 0
	default:
		mode = 1 // ordered
	}
	nt := s.List(mode)
	defer s.End(nt)

	indent := 0
	for {
		ni := s.ListItem()
		indent = len(params[1])
		contentIndent := indent + len(params[2]) + 1
		if strings.HasPrefix(params[3], "[ ] ") || strings.HasPrefix(params[3], "[x] ") {
			s.CheckBox(params[3][1] == 'x')
			params[3] = params[3][4:]
		}
		s.writeLine(params[3], false)
		goto retry

	retry:
		line, ok := s.ReadLine()
		if !ok {
			s.flushNewLine()
			s.End(ni)
			break
		}
		next := syntax.Re.FindStringSubmatch(line)
		if len(next) == 0 {
			if len(line) >= contentIndent && strings.Trim(line[:contentIndent], " \t") == "" {
				s.writeLine(line[contentIndent:], true)
				goto retry
			}
			s.flushNewLine()
			s.End(ni)
			s.Retry()
			break
		}
		params = next
		if len(params[1]) < indent {
			s.flushNewLine()
			s.End(ni)
			s.Retry()
			break
		} else if len(params[1]) > indent {
			s.flushNewLine()
			list(params, s, syntax)
			goto retry
		}
		s.flushNewLine()
		s.End(ni)
	}
}

func code(params []string, s *SyntaxContext, syntax *RegexSyntax) {
	lang := params[1]
	n := s.CodeBlock(lang, params[2])
	defer s.End(n)

	tokenizer := NewTokenizer(lang)
	tokenizer.Code(&DelimiterReader{
		scanner:   s,
		delimiter: []byte("```"),
	})
	for typ, token := tokenizer.Read(); typ != CODE_EOF; typ, token = tokenizer.Read() {
		switch typ {
		case CODE_Keyword:
			s.WriteStyle(token, MarkupData{Attributes: Attributes(".code_key")})
		case CODE_Number:
			s.WriteStyle(token, MarkupData{Attributes: Attributes(".code_num")})
		case CODE_String:
			s.WriteStyle(token, MarkupData{Attributes: Attributes(".code_str")})
		case CODE_Comment:
			s.WriteStyle(token, MarkupData{Attributes: Attributes(".code_comment")})
		case CODE_Ident:
			s.WriteStyle(token, MarkupData{Attributes: Attributes(".code_ident")})
		default:
			s.Write(token)
		}
	}
}

func table(params []string, md *SyntaxContext, syntax *RegexSyntax) {
	align := make(map[int]int, len(params))
	if line, ok := md.ReadLine(); ok {
		t := syntax.Re.FindStringSubmatch(line)
		if len(t) > 0 {
			for i, s := range strings.Split(t[1], "|") {
				if strings.Trim(s, " -:|") != "" {
					md.Retry()
					break
				}
				if strings.HasPrefix(s, ":") {
					align[i] |= 1
				}
				if strings.HasSuffix(s, ":") {
					align[i] |= 2
				}
			}
		} else {
			md.Retry()
		}
	}
	nt := md.Table()
	h := 4
	for {
		text := params[1]
		nr := md.TableRow()
		for i, s := range strings.Split(text, "|") {
			nc := md.TableCell(align[i] | h)
			md.inline(s)
			md.End(nc)
		}
		md.End(nr)
		h = 0

		line, ok := md.ReadLine()
		if !ok {
			break
		}
		text = line
		params = syntax.Re.FindStringSubmatch(text)
		if len(params) < 1 {
			md.Retry()
			break
		}
	}
	md.End(nt)
}

func quote(params []string, md *SyntaxContext, syntax *RegexSyntax) {
	n := md.QuoteBlock()
	defer md.End(n)

	md.inline(params[1] + "\n")
	for {
		text, ok := md.ReadLine()
		if !ok || text == "" {
			break
		}
		params := syntax.Re.FindStringSubmatch(text)
		if len(params) > 0 {
			text = params[1]
		}
		md.inline(text + "\n")
	}
}

var defaultInlineElems []Syntax
var defaultBlockElems []Syntax

func init() {
	defaultInlineElems = []Syntax{
		&SimpleInlineSyntax{"~~", "~~", strike},
		&SimpleInlineSyntax{"**", "**", strong},
		&SimpleInlineSyntax{"*", "*", emphasis},
		&SimpleInlineSyntax{"``", "``", icode},
		&SimpleInlineSyntax{"`", "`", icode},
		&SimpleInlineSyntax{"__", "__", strong},
		&InlineDirectiveSyntax{},
		&ReferenceLinkInlineSyntax{"![", true},
		&ReferenceLinkInlineSyntax{"[", false},
		&LinkInlineSyntax{"["},
		&LinkInlineSyntax{"!["},
		&RegexSyntax{"http", regexp.MustCompile(`^https?:[^\s\"\'\)<>]+`), autolink},
	}
	defaultBlockElems = []Syntax{
		&HeadingSyntax{},
		&RegexSyntax{">", regexp.MustCompile(`^>+\s?(.*)`), quote},
		&RegexSyntax{"```", regexp.MustCompile("^```\\s*(\\w*)(:.*)?$"), code},
		&RegexSyntax{"", regexp.MustCompile(`^\|(.+)\|$`), table},
		&RegexSyntax{"", regexp.MustCompile(`^(\s*)(-|\*|\+|\d+\.)\s(.+)$`), list},
		&RegexSyntax{"", regexp.MustCompile(`^([-_]\s?){3,}$`), hr},
	}
}
