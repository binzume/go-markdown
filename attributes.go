package markdown

import "strings"

// Attributes is the contents of a Markdown attribute block. It may be given
// with or without the surrounding braces, for example "#heading .important"
// or "{#heading .important}".
type Attributes string

// GetIdAndClassName returns the ID and CSS class name represented by a
// Markdown attribute block. It returns empty strings when the attributes are
// invalid or contain unsupported fields.
func (a Attributes) GetIdAndClassName() (id, className string) {
	id, className, _ = a.values()
	return id, className
}

func (a Attributes) markupData() (MarkupData, bool) {
	if _, _, ok := a.values(); !ok {
		return MarkupData{}, false
	}
	return MarkupData{Attributes: a}, true
}

func (a Attributes) values() (id, className string, ok bool) {
	attributes := strings.TrimSpace(string(a))
	if strings.HasPrefix(attributes, "{") && strings.HasSuffix(attributes, "}") {
		attributes = strings.TrimSpace(attributes[1 : len(attributes)-1])
	}

	fields := strings.Fields(attributes)
	if len(fields) == 0 {
		return "", "", false
	}

	classes := make([]string, 0, len(fields))
	for _, field := range fields {
		switch {
		case strings.HasPrefix(field, "#") && len(field) > 1 && id == "":
			id = field[1:]
		case strings.HasPrefix(field, ".") && len(field) > 1:
			classes = append(classes, field[1:])
		default:
			return "", "", false
		}
	}
	return id, strings.Join(classes, " "), true
}
