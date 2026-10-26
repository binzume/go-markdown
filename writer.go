package markdown

// MarkupData contains optional attributes for markup elements.
//
// Additional fields can be added here as more markup attributes are supported.
type MarkupData struct {
	Attributes Attributes
	Data       any
}

type DocWriter interface {
	Block(data MarkupData) int
	Span(data MarkupData) int
	Heading(level int, data MarkupData) int
	Link(url string, title string, data MarkupData) int
	Image(url string, title, alt string, data MarkupData) int
	Hr() int
	Strike() int
	Emphasis() int
	Strong() int
	Code() int
	Paragraph() int
	List(mode int) int
	ListItem() int
	Table() int
	TableRow() int
	TableCell(flags int) int
	CheckBox(checked bool) int
	QuoteBlock() int
	CodeBlock(lang string, title string) int
	End(lv int)
	Write(text string)
	WriteStyle(text string, data MarkupData)
	Close()
}
