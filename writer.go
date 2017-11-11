package markdown

type DocWriter interface {
	Heading(text string, level int) int
	Link(url string, title string, options int) int
	Image(url string, title string, options int) int
	Hr() int
	Strike() int
	Emphasis() int
	Strong() int
	Bold() int
	Code() int
	Paragraph() int
	List(mode int) int
	ListItem() int
	Table() int
	TableRow() int
	TableCell(flags int) int
	QuoteBlock() int
	CodeBlock(lang string, title string) int
	End(lv int)
	Write(text string)
	WriteStyle(text string, className string, color string, flags int)
	Close()
}
