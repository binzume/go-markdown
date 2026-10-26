package directives

import (
	"bufio"
	"strings"

	markdown "github.com/binzume/go-markdown"
)

// RegisterNote adds Qiita-compatible note directives to md.
func RegisterNote(md *markdown.Markdown) {
	md.RegisterDirective("note", func(d *markdown.DirectiveContext) error {
		level := "info"
		if fields := strings.Fields(d.Argument); len(fields) > 0 {
			level = fields[0]
		}

		n := d.Writer.Block(markdown.MarkupData{Attributes: markdown.Attributes(".note-" + level)})
		err := md.Convert(bufio.NewScanner(d.Content), d.Writer)
		d.Writer.End(n)
		return err
	})
}
