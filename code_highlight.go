package markdown

import (
	"strings"
	"text/scanner"
)

var RubyKeywords = strings.Split("class,def,if,else,unless,do,next,begin,end,ensure,new,attr_accessor,return,require,require_relative", ",")
var PerlKeywords = strings.Split("for,foreach,if,else,elsif,do,while,next,last,return,sub,my,qw,local,require,use", ",")
var CKeywords = strings.Split("void,int,char,float,double,long,short,signed,unsigned,volatile,"+
	"static,const,auto,for,if,else,do,while,continue,break,return,switch,case,default,typedef,enum,struct", ",")
var CppKeywords = strings.Split("class,public,private,protected,namespace,using,bool,new,delete", ",")
var GoKeywords = strings.Split("package,import,var,type,for,if,else,continue,break,return,func,switch,case,default,int,string,map,float", ",")

var PhpKeywords = strings.Split("for,foreach,if,else,elseif,do,while,continue,break,new,class,return,catch,try,global,public,private,function,switch,case", ",")

var RustKeywords = strings.Split("fn,let,enum,mod,struct,trait,type,use,impl,box,crate,where,true,false,self,super,"+
	"if,else,match,for,loop,while,break,continue,return,as,in,const,static,pub,mut,move,ref,unsafe,extern", ",")

var Keywords = map[string][]string{
	"go":   GoKeywords,
	"ruby": RubyKeywords,
	"rb":   RubyKeywords,
	"rust": RustKeywords,
	"php":  PhpKeywords,
	"perl": PerlKeywords,
	"pl":   PerlKeywords,
	"c":    CKeywords,
	"cpp":  append(CKeywords, CppKeywords...),
	"js":   append(CKeywords, strings.Split("var,function,new", ",")...),
}

const CODE_EOF = -1
const CODE_Keyword = 0
const CODE_Number = 1
const CODE_String = 2
const CODE_Ident = 4
const CODE_Comment = 9
const CODE_UNKNOWN = 100

type Tokenizer struct {
	Lang     string
	Mode     int
	keywords map[string]int
	scanner  scanner.Scanner
}

func NewTokenizer(lang string) *Tokenizer {
	t := &Tokenizer{Lang: lang, keywords: make(map[string]int)}

	for _, k := range Keywords[lang] {
		t.keywords[k] = 1
	}

	return t
}

func (t *Tokenizer) Code(code string) {
	t.scanner.Init(strings.NewReader(code))
	t.scanner.Mode = scanner.GoTokens ^ scanner.SkipComments
	t.scanner.Whitespace = 0
	t.scanner.Error = func(s *scanner.Scanner, msg string) {}
}

func (t *Tokenizer) Read() (int, string) {
	tok := t.scanner.Scan()
	s := t.scanner.TokenText()
	switch tok {
	case scanner.EOF:
		return CODE_EOF, s
	case scanner.Ident:
		if t.keywords[s] > 0 {
			return CODE_Keyword, s
		}
		return CODE_Ident, s
	case scanner.String, scanner.Char:
		return CODE_String, s
	case scanner.Comment:
		return CODE_Comment, s
	case scanner.Float, scanner.Int:
		return CODE_Number, s
	default:
		return CODE_UNKNOWN, s
	}
}
