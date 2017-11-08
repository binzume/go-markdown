package main

type DocWriter interface {
	Heading(text string, level int) int
	Link(url string, options int) int
	Image(url string, options int) int
	Strike() int
	Bold() int
	List() int
	ListItem() int
	Table() int
	TableRow() int
	TableCell() int
	CodeBlock(lang string) int
	End(lv int)
	Write(text string)
	WriteStyle(text string, className string, color string, flags int)
	Close()
}
