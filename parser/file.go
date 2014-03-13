package parser

import (
	"fmt"
)

type Idx int

type Position struct {
	Name   string
	Line   int
	Column int
}

func (self *Position) String() string {
	name := "(anonymous)"
	if self.Name != "" {
		name = self.Name
	}
	return fmt.Sprintf("%s:%d:%d:", name, self.Line, self.Column)
}
