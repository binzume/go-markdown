package markdown

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
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

func (m *SimpleInlineSyntax) Prefix() string {
	return m.Start
}

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

func (m *RegexSyntax) Prefix() string {
	return m.PrefixStr
}

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

func (m *HeadingSyntax) Prefix() string {
	return "#"
}

func (m *HeadingSyntax) TryParse(line string) (int, []string) {
	level := 0
	for level < len(line) && level < 4 && line[level] == '#' {
		level++
	}
	if level == 0 {
		return -1, nil
	}

	content := line[level:]
	content = strings.TrimLeft(content, " \t\r\f")
	return len(line), []string{line[:level], content}
}

func (m *HeadingSyntax) Render(params []string, s *SyntaxContext) {
	s.Heading(params[1], len(params[0]))
}

type LinkInlineSyntax struct {
	Start string
}

func (m *LinkInlineSyntax) Prefix() string {
	return m.Start
}

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
	return p2 + 1, []string{line[len(m.Start):p1], line[p1+2 : p2]}
}

func (m *LinkInlineSyntax) Render(params []string, md *SyntaxContext) {
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

type ReferenceLinkInlineSyntax struct {
	Start string
	Image bool
}

func (m *ReferenceLinkInlineSyntax) Prefix() string {
	return m.Start
}

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
	n := md.Link(params[0], "", 0)
	md.Write(params[0])
	md.End(n)
}

func heading(params []string, md *SyntaxContext, syntax *RegexSyntax) {
	md.Heading(params[2], len(params[1]))
}

func hr(params []string, md *SyntaxContext, syntax *RegexSyntax) {
	md.Hr()
}

func dummy(params []string, md *SyntaxContext, syntax *RegexSyntax) {
}

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
		if strings.HasPrefix(params[3], "[ ] ") || strings.HasPrefix(params[3], "[x] ") {
			s.CheckBox(params[3][1] == 'x')
			params[3] = params[3][4:]
		}
		s.inline(params[3])
		s.End(ni)
	retry:
		line, ok := s.ReadLine()
		if !ok {
			break
		}
		params = syntax.Re.FindStringSubmatch(line)
		if len(params) < 1 || len(params[1]) < indent {
			s.Retry()
			break
		} else if len(params[1]) > indent {
			list(params, s, syntax)
			goto retry
		}
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
		if !ok {
			break
		}
		if text == "" {
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

var directiveStart = regexp.MustCompile(`^(:{3,})\s*([A-Za-z][A-Za-z0-9_-]*)(?:\s+(.*?))?\s*$`)

func init() {
	defaultInlineElems = []Syntax{
		&SimpleInlineSyntax{"~~", "~~", strike},
		&SimpleInlineSyntax{"**", "**", strong},
		&SimpleInlineSyntax{"*", "*", emphasis},
		&SimpleInlineSyntax{"``", "``", icode},
		&SimpleInlineSyntax{"`", "`", icode},
		&SimpleInlineSyntax{"__", "__", strong},
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

// DirectiveHandler handles a fenced block directive.
type DirectiveHandler func(*DirectiveContext) error

// DirectiveContext is passed to a registered DirectiveHandler.
type DirectiveContext struct {
	Name     string
	Argument string
	Content  io.Reader
	Writer   DocWriter
}

// Markdown config.
type Markdown struct {
	inlineElems   []Syntax
	blockElems    []Syntax
	inlineCharMap map[byte]bool
	directives    map[string]DirectiveHandler
}

// NewMarkdown returns *Markdown
func NewMarkdown() *Markdown {
	m := make(map[byte]bool)
	for _, markup := range defaultInlineElems {
		m[markup.Prefix()[0]] = true
	}
	return &Markdown{defaultInlineElems, defaultBlockElems, m, make(map[string]DirectiveHandler)}
}

// RegisterDirective registers a block directive handler.
func (m *Markdown) RegisterDirective(name string, handler DirectiveHandler) {
	m.directives[name] = handler
}

// SyntaxContext provides the parser and writer state available to a Syntax.
type SyntaxContext struct {
	*Markdown
	*bufio.Scanner
	DocWriter
	lineText   string
	lineBytes  []byte
	retry      bool
	references map[string]reference
}

func (s *SyntaxContext) ReadLine() (string, bool) {
	if s.retry {
		s.retry = false
		return s.lineText, true
	}
	if !s.Scanner.Scan() {
		return "", false
	}
	s.lineText = s.Scanner.Text()
	s.lineBytes = s.Scanner.Bytes()
	return s.lineText, true
}

func (s *SyntaxContext) Err() error {
	return s.Scanner.Err()
}

func (s *SyntaxContext) Retry() {
	s.retry = true
}

type reference struct {
	URL   string
	Title string
}

func normalizeReferenceLabel(label string) string {
	return strings.ToLower(strings.Join(strings.Fields(label), " "))
}

func parseReferenceDefinition(line string) (string, reference, bool) {
	line = strings.TrimRight(line, " \t")
	trimmed := strings.TrimLeft(line, " ")
	if len(line)-len(trimmed) > 3 || !strings.HasPrefix(trimmed, "[") {
		return "", reference{}, false
	}

	end := strings.Index(trimmed, "]:")
	if end <= 1 {
		return "", reference{}, false
	}
	label := trimmed[1:end]
	rest := strings.TrimSpace(trimmed[end+2:])
	if label == "" || rest == "" {
		return "", reference{}, false
	}

	var destination string
	if rest[0] == '<' {
		close := strings.IndexByte(rest[1:], '>')
		if close < 0 {
			return "", reference{}, false
		}
		close++
		destination = rest[1:close]
		rest = strings.TrimSpace(rest[close+1:])
	} else {
		end := strings.IndexAny(rest, " \t")
		if end < 0 {
			destination = rest
			rest = ""
		} else {
			destination = rest[:end]
			rest = strings.TrimSpace(rest[end:])
		}
	}
	if destination == "" {
		return "", reference{}, false
	}

	title := ""
	if rest != "" {
		if len(rest) < 2 {
			return "", reference{}, false
		}
		open, close := rest[0], rest[len(rest)-1]
		if (open != '"' && open != '\'' && open != '(') ||
			(open == '"' && close != '"') ||
			(open == '\'' && close != '\'') ||
			(open == '(' && close != ')') {
			return "", reference{}, false
		}
		title = rest[1 : len(rest)-1]
	}

	return normalizeReferenceLabel(label), reference{URL: destination, Title: title}, true
}

func (s *SyntaxContext) addReference(label string, ref reference) {
	if s.references == nil {
		s.references = make(map[string]reference)
	}
	s.references[label] = ref
}

// WriteReferenceLink writes a stored reference link using the current DocWriter.
// It returns false when label has not been defined.
func (s *SyntaxContext) WriteReferenceLink(text, label string) bool {
	ref, ok := s.references[normalizeReferenceLabel(label)]
	if !ok {
		return false
	}
	n := s.Link(ref.URL, ref.Title, 0)
	s.inline(text)
	s.End(n)
	return true
}

// WriteReferenceImage writes a stored reference image using the current DocWriter.
// It returns false when label has not been defined.
func (s *SyntaxContext) WriteReferenceImage(alt, label string) bool {
	ref, ok := s.references[normalizeReferenceLabel(label)]
	if !ok {
		return false
	}
	n := s.Image(ref.URL, ref.Title, alt, 0)
	s.End(n)
	return true
}

func parseDirectiveStart(line string) (name string, argument string, fence int, ok bool) {
	match := directiveStart.FindStringSubmatch(line)
	if len(match) == 0 {
		return "", "", 0, false
	}

	return match[2], strings.TrimSpace(match[3]), len(match[1]), true
}

type DelimiterReader struct {
	scanner   *SyntaxContext
	delimiter []byte
	buf       []byte
	done      bool
	found     bool
}

func (r *DelimiterReader) Read(p []byte) (n int, err error) {
	if r.done {
		if err := r.scanner.Err(); err != nil {
			return 0, err
		}
		return 0, io.EOF
	}
	if len(r.buf) == 0 {
		line, ok := r.scanner.ReadLine()
		if !ok {
			r.done = true
			if err := r.scanner.Err(); err != nil {
				return 0, err
			}
			return 0, io.EOF
		}
		if bytes.Equal(bytes.TrimSpace([]byte(line)), r.delimiter) {
			r.done = true
			r.found = true
			return 0, io.EOF
		}
		r.buf = r.scanner.lineBytes
		r.buf = append(r.buf, '\n')
	}
	l := copy(p, r.buf)
	r.buf = r.buf[l:]
	return l, nil
}

func (s *SyntaxContext) inline(text string) {
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
					l, params := markup.TryParse(text[pos:])
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

func (s *SyntaxContext) block() error {
	writer := s.DocWriter
	para := 0
	for {
		text, ok := s.ReadLine()
		if !ok {
			break
		}
		if label, ref, ok := parseReferenceDefinition(text); ok {
			s.addReference(label, ref)
			continue
		}
		if name, argument, fence, ok := parseDirectiveStart(text); ok {
			if para != 0 {
				writer.End(para)
				para = 0
			}
			writer.Write("\n")

			closingFence := strings.Repeat(":", fence)
			content := &DelimiterReader{
				scanner:   s,
				delimiter: []byte(closingFence),
			}
			consumed := false
			if handler := s.directives[name]; handler != nil {
				ctx := &DirectiveContext{
					Name:     name,
					Argument: argument,
					Content:  content,
					Writer:   writer,
				}
				if err := handler(ctx); err != nil {
					return fmt.Errorf("directive %q: %w", name, err)
				}
			} else {
				n := writer.Block("")
				child := &SyntaxContext{
					Markdown:   s.Markdown,
					Scanner:    bufio.NewScanner(content),
					DocWriter:  writer,
					references: s.references,
				}
				err := child.block()
				writer.End(n)
				if err != nil {
					return fmt.Errorf("directive %q: %w", name, err)
				}
				consumed = true
			}
			if !consumed {
				if _, err := io.Copy(io.Discard, content); err != nil {
					return err
				}
			}
			if !content.found {
				return fmt.Errorf("unclosed directive %q", name)
			}
			continue
		}
		for _, matcher := range s.blockElems {
			if prefix := matcher.Prefix(); prefix != "" && !strings.HasPrefix(text, prefix) {
				continue
			}
			l, params := matcher.TryParse(text)
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
		} else {
			s.Write("\n")
		}
		s.inline(text)
	}
	return s.Err()
}

// Convert md to html.
func Convert(scanner0 *bufio.Scanner, writer DocWriter) error {
	return NewMarkdown().Convert(scanner0, writer)
}

// Convert renders Markdown with the directives registered on m.
func (m *Markdown) Convert(scanner0 *bufio.Scanner, writer DocWriter) error {
	state := &SyntaxContext{Markdown: m, Scanner: scanner0, DocWriter: writer, references: make(map[string]reference)}
	return state.block()
}
