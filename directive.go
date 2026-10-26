package markdown

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// DirectiveHandler handles a directive.
type DirectiveHandler func(*DirectiveContext) error

// DirectiveContext is passed to a registered DirectiveHandler.
type DirectiveContext struct {
	Name     string
	Argument string
	Inline   bool
	Content  io.Reader
	Writer   DocWriter
}

// RegisterDirective registers a directive handler.
func (m *Markdown) RegisterDirective(name string, handler DirectiveHandler) {
	m.directives[name] = handler
}

type InlineDirectiveSyntax struct{}

func (m *InlineDirectiveSyntax) Prefix() string {
	return ":"
}

func (m *InlineDirectiveSyntax) TryParse(line string) (int, []string) {
	if !strings.HasPrefix(line, ":") {
		return -1, nil
	}

	pos := 1
	if pos >= len(line) || !isDirectiveNameStart(line[pos]) {
		return -1, nil
	}
	for pos < len(line) && isDirectiveNameChar(line[pos]) {
		pos++
	}
	name := line[1:pos]
	if pos >= len(line) || line[pos] != '[' {
		return -1, nil
	}

	bodyEnd := strings.IndexByte(line[pos+1:], ']')
	if bodyEnd < 0 {
		return -1, nil
	}
	bodyEnd += pos + 1
	if bodyEnd+1 >= len(line) || line[bodyEnd+1] != '{' {
		return -1, nil
	}

	argumentEnd := strings.IndexByte(line[bodyEnd+2:], '}')
	if argumentEnd < 0 {
		return -1, nil
	}
	argumentEnd += bodyEnd + 2
	return argumentEnd + 1, []string{
		name,
		line[pos+1 : bodyEnd],
		line[bodyEnd+2 : argumentEnd],
	}
}

func (m *InlineDirectiveSyntax) Render(params []string, s *SyntaxContext) {
	s.renderInlineDirective(params[0], params[1], params[2])
}

func isDirectiveNameStart(c byte) bool {
	return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z'
}

func isDirectiveNameChar(c byte) bool {
	return isDirectiveNameStart(c) || c >= '0' && c <= '9' || c == '_' || c == '-'
}

func (s *SyntaxContext) renderInlineDirective(name, body, argument string) {
	if handler := s.directives[name]; handler != nil {
		ctx := &DirectiveContext{
			Name:     name,
			Argument: argument,
			Inline:   true,
			Content:  strings.NewReader(body),
			Writer:   s.DocWriter,
		}
		if err := handler(ctx); err != nil && s.err == nil {
			s.err = fmt.Errorf("inline directive %q: %w", name, err)
		}
		return
	}

	n := s.Span(MarkupData{})
	s.inline(body)
	s.End(n)
}

var directiveStart = regexp.MustCompile(`^(:{3,})\s*([A-Za-z][A-Za-z0-9_-]*)(?:\s+(.*?))?\s*$`)

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

func (s *SyntaxContext) directive(name, argument string, fence int) error {
	content := &DelimiterReader{
		scanner:   s,
		delimiter: []byte(strings.Repeat(":", fence)),
	}
	consumed := false
	if handler := s.directives[name]; handler != nil {
		ctx := &DirectiveContext{
			Name:     name,
			Argument: argument,
			Content:  content,
			Writer:   s.DocWriter,
		}
		if err := handler(ctx); err != nil {
			return fmt.Errorf("directive %q: %w", name, err)
		}
	} else {
		n := s.Block(MarkupData{})
		child := &SyntaxContext{
			Markdown:   s.Markdown,
			Scanner:    bufio.NewScanner(content),
			DocWriter:  s.DocWriter,
			references: s.references,
		}
		err := child.block()
		s.End(n)
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
	return nil
}
