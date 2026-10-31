package markdown

import (
	"bufio"
	"io"
	"strings"
)

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

// SyntaxContext provides the parser and writer state available to a Syntax.
type SyntaxContext struct {
	*Markdown
	*bufio.Scanner
	DocWriter
	lineText   string
	lineBytes  []byte
	retry      bool
	hardBreak  bool
	references map[string]reference
	err        error
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

func (s *SyntaxContext) writeLine(text string, continued bool) {
	if continued {
		if s.hardBreak {
			s.NewLine()
		} else {
			s.Write("\n")
		}
	}

	hardBreak := strings.HasSuffix(text, "  ")
	if hardBreak {
		text = strings.TrimRight(text, " ")
	}
	s.inline(text)
	s.hardBreak = hardBreak
}

func (s *SyntaxContext) flushNewLine() {
	if s.hardBreak {
		s.NewLine()
		s.hardBreak = false
	}
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
	n := s.Link(ref.URL, ref.Title, MarkupData{})
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
	n := s.Image(ref.URL, ref.Title, alt, MarkupData{})
	s.End(n)
	return true
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
				s.flushNewLine()
				writer.End(para)
				para = 0
			}
			writer.Write("\n")

			if err := s.directive(name, argument, fence); err != nil {
				return err
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
					s.flushNewLine()
					writer.End(para)
					para = 0
				}
				writer.Write("\n")
				matcher.Render(params, s)
				if s.err != nil {
					return s.err
				}
				text = ""
				break
			}
		}
		if text == "" {
			if para != 0 {
				s.flushNewLine()
				writer.End(para)
				para = 0
			}
			continue
		}
		continued := para != 0
		if para == 0 {
			para = writer.Paragraph()
		}
		s.writeLine(text, continued)
		if s.err != nil {
			return s.err
		}
	}
	s.flushNewLine()
	return s.Err()
}

// Convert renders Markdown read from reader with writer.
func Convert(reader io.Reader, writer DocWriter) error {
	return NewMarkdown().Convert(reader, writer)
}

// Convert renders Markdown with the directives registered on m.
func (m *Markdown) Convert(reader io.Reader, writer DocWriter) error {
	scanner := bufio.NewScanner(reader)
	state := &SyntaxContext{Markdown: m, Scanner: scanner, DocWriter: writer, references: make(map[string]reference)}
	return state.block()
}
