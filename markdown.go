package markdown

import (
	"bufio"
	"regexp"
	"strings"
)

type Matcher interface {
	Prefix() string
	TryMatch(line string) (int, []string)
	Render(params []string, s *state)
}

type SimpleInlineMatcher struct {
	Start      string
	End        string
	RenderFunc func(content string, s *state, matcher *SimpleInlineMatcher)
}

func (m *SimpleInlineMatcher) Prefix() string {
	return m.Start
}

func (m *SimpleInlineMatcher) TryMatch(line string) (int, []string) {
	line = line[len(m.Start):]
	p := strings.Index(line, m.End)
	if p <= 0 {
		return -1, nil
	}
	return p + len(m.End) + len(m.Start), []string{line[:p]}
}

func (m *SimpleInlineMatcher) Render(params []string, s *state) {
	m.RenderFunc(params[0], s, m)
}

type RegexMatcher struct {
	PrefixStr  string
	Re         *regexp.Regexp
	RenderFunc func(matches []string, s *state, matcher *RegexMatcher)
}

func (m *RegexMatcher) Prefix() string {
	return m.PrefixStr
}

func (m *RegexMatcher) TryMatch(text string) (int, []string) {
	match := m.Re.FindStringSubmatch(text)
	if len(match) < 1 {
		return -1, nil
	}
	return len(match[0]), match
}

func (m *RegexMatcher) Render(params []string, s *state) {
	m.RenderFunc(params, s, m)
}

type LinkInlineMatcher struct {
	Start string
}

func (m *LinkInlineMatcher) Prefix() string {
	return m.Start
}

func (m *LinkInlineMatcher) TryMatch(line string) (int, []string) {
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
	return p2 + 1, []string{line[len(m.Start):p1], line[p1+2 : p2]}
}

func (m *LinkInlineMatcher) Render(params []string, md *state) {
	url := strings.Split(params[1], " ")
	var title string
	if len(url) > 1 {
		title = strings.Trim(url[1], "\" ")
	}
	if m.Start == "![" {
		n := md.Image(url[0], title, params[0], 0)
		md.End(n)
	} else {
		n := md.Link(url[0], title, 0)
		md.inline(params[0])
		md.End(n)
	}
}

func strike(text string, md *state, markup *SimpleInlineMatcher) {
	n := md.Strike()
	md.inline(text)
	md.End(n)
}

func strong(text string, md *state, markup *SimpleInlineMatcher) {
	n := md.Strong()
	md.inline(text)
	md.End(n)
}

func emphasis(text string, md *state, markup *SimpleInlineMatcher) {
	n := md.Emphasis()
	md.inline(text)
	md.End(n)
}

func icode(text string, md *state, markup *SimpleInlineMatcher) {
	n := md.Code()
	md.Write(text)
	md.End(n)
}

func autolink(params []string, md *state, markup *RegexMatcher) {
	n := md.Link(params[0], "", 0)
	md.Write(params[0])
	md.End(n)
}

func heading(params []string, md *state, markup *RegexMatcher) {
	md.Heading(params[2], len(params[1]))
}

func hr(params []string, md *state, markup *RegexMatcher) {
	md.Hr()
}

func dummy(params []string, md *state, markup *RegexMatcher) {
}

func list(params []string, s *state, markup *RegexMatcher) {
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
		if strings.HasPrefix(params[3], "[ ] ") || strings.HasPrefix(params[3], "[x] ") {
			s.CheckBox(params[3][1] == 'x')
			params[3] = params[3][4:]
		}
		s.inline(params[3])
		s.End(ni)
	retry:
		if !s.Scan() {
			break
		}
		text := s.Text()
		params = markup.Re.FindStringSubmatch(text)
		if len(params) < 1 || len(params[1]) < indent {
			s.Retry()
			break
		} else if len(params[1]) > indent {
			list(params, s, markup)
			goto retry
		}
	}
}

func code(params []string, s *state, markup *RegexMatcher) {
	lang := params[1]
	n := s.CodeBlock(lang, params[2])
	defer s.End(n)

	tokenizer := NewTokenizer(lang)

	for s.Scan() {
		text := s.Text()
		m := markup.Re.FindStringSubmatch(text)
		if len(m) > 0 {
			break
		}
		tokenizer.Code(text)
		for typ, token := tokenizer.Read(); typ != CODE_EOF; typ, token = tokenizer.Read() {
			switch typ {
			case CODE_Keyword:
				s.WriteStyle(token, "code_key", "", 0)
			case CODE_Number:
				s.WriteStyle(token, "code_num", "", 0)
			case CODE_String:
				s.WriteStyle(token, "code_str", "", 0)
			case CODE_Comment:
				s.WriteStyle(token, "code_comment", "", 0)
			case CODE_Ident:
				s.WriteStyle(token, "code_ident", "", 0)
			default:
				s.Write(token)
			}
		}
		s.Write("\n")
	}
}

func table(params []string, md *state, markup *RegexMatcher) {
	align := make(map[int]int, len(params))
	if md.Scan() {
		t := markup.Re.FindStringSubmatch(md.Text())
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

		if !md.Scan() {
			break
		}
		text = md.Text()
		params = markup.Re.FindStringSubmatch(text)
		if len(params) < 1 {
			md.Retry()
			break
		}
	}
	md.End(nt)
}

func quote(params []string, md *state, markup *RegexMatcher) {
	n := md.QuoteBlock()
	defer md.End(n)

	md.inline(params[1] + "\n")
	for md.Scan() {
		text := md.Text()
		if text == "" {
			break
		}
		params := markup.Re.FindStringSubmatch(text)
		if len(params) > 0 {
			text = params[1]
		}
		md.inline(text + "\n")
	}
}

func pluginBlock(params []string, md *state, markup *RegexMatcher) {
	// TODO
	for md.Scan() {
		text := md.Text()
		if text == "}" {
			break
		}
	}
}

var defaultInlineElems []Matcher
var defaultBlockElems []Matcher

func init() {
	defaultInlineElems = []Matcher{
		&SimpleInlineMatcher{"~~", "~~", strike},
		&SimpleInlineMatcher{"**", "**", strong},
		&SimpleInlineMatcher{"*", "*", emphasis},
		&SimpleInlineMatcher{"``", "``", icode},
		&SimpleInlineMatcher{"`", "`", icode},
		&SimpleInlineMatcher{"__", "__", strong},
		&LinkInlineMatcher{"["},
		&LinkInlineMatcher{"!["},
		&RegexMatcher{"http", regexp.MustCompile(`^https?:[^\s\"\'\)<>]+`), autolink},
	}
	defaultBlockElems = []Matcher{
		&RegexMatcher{"#", regexp.MustCompile(`^(#{1,4})\s*(.*)`), heading},
		&RegexMatcher{">", regexp.MustCompile(`^>+\s?(.*)`), quote},
		&RegexMatcher{"```", regexp.MustCompile("^```\\s*(\\w*)(:.*)?$"), code},
		&RegexMatcher{"", regexp.MustCompile(`^\|(.+)\|$`), table},
		&RegexMatcher{"", regexp.MustCompile(`^(\s*)(-|\*|\+|\d+\.)\s(.+)$`), list},
		&RegexMatcher{"", regexp.MustCompile(`^([-_]\s?){3,}$`), hr},
		&RegexMatcher{"[", regexp.MustCompile(`^\[([^\]]+)\]:\s+([^\s]+)\s+(.*)`), dummy}, // fixme
		&RegexMatcher{"&", regexp.MustCompile(`^&(\w+)[{]*$`), pluginBlock},
	}
}

// Markdown config.
type Markdown struct {
	inlineElems   []Matcher
	blockElems    []Matcher
	inlineCharMap map[byte]bool
}

// NewMarkdown returns *Markdown
func NewMarkdown() *Markdown {
	m := make(map[byte]bool)
	for _, markup := range defaultInlineElems {
		m[markup.Prefix()[0]] = true
	}
	return &Markdown{defaultInlineElems, defaultBlockElems, m}
}

type state struct {
	*Markdown
	*bufio.Scanner
	DocWriter
	retry bool
}

func (s *state) Scan() bool {
	if s.retry {
		s.retry = false
		return true
	}
	return s.Scanner.Scan()
}

func (s *state) Retry() {
	s.retry = true
}

func (s *state) inline(text string) {
	writer := s.DocWriter
	for pos := 0; pos < len(text); pos++ {
		// TODO more fast.
		if s.inlineCharMap[text[pos]] || true {
			if pos > 0 && text[pos-1] == '\\' {
				// Escaped
				writer.Write(text[:pos-1])
				text = text[pos:]
				pos = 1
				continue
			}
			for _, markup := range s.inlineElems {
				if strings.HasPrefix(text[pos:], markup.Prefix()) {
					l, params := markup.TryMatch(text[pos:])
					if l > 0 {
						writer.Write(text[:pos])
						markup.Render(params, s)
						text = text[(pos + l):]
						pos = 0
					}
				}

			}
		}
	}
	writer.Write(text)
}

func (s *state) block() {
	writer := s.DocWriter
	para := 0
	for s.Scan() {
		text := s.Text()
		for _, matcher := range s.blockElems {
			l, params := matcher.TryMatch(text)
			if l > 0 {
				if para != 0 {
					writer.End(para)
					para = 0
				}
				writer.Write("\n")
				matcher.Render(params, s)
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
		s.inline(text)
	}
}

// Convert md to html.
func Convert(scanner0 *bufio.Scanner, writer DocWriter) error {
	state := &state{NewMarkdown(), scanner0, writer, false}
	state.block()
	// NewMarkdown().block(&Source{scanner: scanner0}, writer)
	return scanner0.Err()
}
