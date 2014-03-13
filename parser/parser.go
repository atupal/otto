package parser

import (
	"github.com/robertkrimen/otto/parser/token"
)

type _parser struct {
	filename string
	source   string
	length   int
	base     Idx

	chr       rune
	chrOffset int
	offset    int

	token             token.Token
	literal           string
	idx               Idx
	scope             *_scope
	insertSemicolon   bool
	implicitSemicolon bool

	errors []error

	recover struct {
		idx   Idx
		count int
	}
}

func newParser(filename, source string) *_parser {
	return &_parser{
		chr:    ' ', // This is set so we can start scanning by skipping whitespace
		source: source,
		length: len(source),
		base:   Idx(1),
	}
}

func Parse(filename, source string) (*Program, error) {
	parser := newParser(filename, source)
	return parser.parse()
}

func (self *_parser) slice(idx0, idx1 Idx) string {
	from := int(idx0 - self.base)
	to := int(idx1 - self.base)
	if from >= 0 && to <= len(self.source) {
		return self.source[from:to]
	}

	return ""
}

func (self *_parser) parse() (*Program, error) {
	self.next()
	program := self.parseProgram()
	if len(self.errors) > 0 {
		return program, self.errors[0]
	}
	return program, nil
}

func (self *_parser) next() {
	self.token, self.literal, self.idx = self.scan()
}

func (self *_parser) optionalSemicolon() {
	if self.token == token.SEMICOLON {
		self.next()
		return
	}

	if self.implicitSemicolon {
		self.implicitSemicolon = false
		return
	}

	if self.token != token.EOF && self.token != token.RIGHT_BRACE {
		self.expect(token.SEMICOLON)
	}
}

func (self *_parser) semicolon() {
	if self.token != token.RIGHT_PARENTHESIS && self.token != token.RIGHT_BRACE {
		if self.implicitSemicolon {
			self.implicitSemicolon = false
			return
		}

		self.expect(token.SEMICOLON)
	}
}

func (self *_parser) idxOf(offset int) Idx {
	return self.base + Idx(offset)
}

func (self *_parser) expect(value token.Token) Idx {
	idx := self.idx
	if self.token != value {
		self.errorUnexpectedToken(self.token)
	}
	self.next()
	return idx
}

func lineCount(source string) (int, int) {
	line, last := 0, -1
	pair := false
	for index, chr := range source {
		switch chr {
		case '\r':
			line += 1
			last = index
			pair = true
			continue
		case '\n':
			if !pair {
				line += 1
			}
			last = index
		case '\u2028', '\u2029':
			line += 1
			last = index + 2
		}
		pair = false
	}
	return line, last
}

func (self *_parser) position(idx Idx) *Position {
	position := &Position{}
	offset := int(idx - self.base)
	source := self.source[:offset]
	position.Name = self.filename
	line, last := lineCount(source)
	position.Line = 1 + line
	if last >= 0 {
		position.Column = offset - last
	} else {
		position.Column = 1 + len(source)
	}

	return position
}
