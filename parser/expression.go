package parser

import (
	"regexp"

	"github.com/robertkrimen/otto/parser/token"
)

func (self *_parser) parseIdentifier() Identifier {
	literal := self.literal
	idx := self.idx
	self.next()
	return Identifier{
		Name: literal,
		Idx:  idx,
	}
}

func (self *_parser) parsePrimaryExpression() Expression {
	literal := self.literal
	idx := self.idx
	switch self.token {
	case token.IDENTIFIER:
		self.next()
		return &Identifier{
			Name: literal,
			Idx:  idx,
		}
	case token.NULL:
		self.next()
		return &NullLiteral{
			Idx:     idx,
			Literal: literal,
		}
	case token.BOOLEAN:
		self.next()
		value := false
		switch literal {
		case "true":
			value = true
		case "false":
			value = false
		default:
			self.error(idx, "Illegal boolean literal")
		}
		return &BooleanLiteral{
			Idx:     idx,
			Literal: literal,
			Value:   value,
		}
	case token.STRING:
		self.next()
		value := parseStringLiteral(literal[1 : len(literal)-1])
		return &StringLiteral{
			Idx:     idx,
			Literal: literal,
			Value:   value,
		}
	case token.NUMBER:
		self.next()
		value, err := parseNumberLiteral(literal)
		if err != nil {
			self.error(idx, err.Error())
			value = 0
		}
		return &NumberLiteral{
			Idx:     idx,
			Literal: literal,
			Value:   value,
		}
	case token.SLASH, token.QUOTIENT_ASSIGN:
		return self.parseRegExpLiteral()
	case token.LEFT_BRACE:
		return self.parseObjectLiteral()
	case token.LEFT_BRACKET:
		return self.parseArrayLiteral()
	case token.LEFT_PARENTHESIS:
		self.expect(token.LEFT_PARENTHESIS)
		expression := self.parseExpression()
		self.expect(token.RIGHT_PARENTHESIS)
		return expression
	case token.THIS:
		self.next()
		return &ThisExpression{
			Idx: idx,
		}
	case token.FUNCTION:
		return self.parseFunction(false)
	}

	self.errorUnexpectedToken(self.token)
	self.nextStatement()
	return &BadExpression{From: idx, To: self.idx}
}

func (self *_parser) parseRegExpLiteral() *RegExpLiteral {

	offset := self.chrOffset - 1 // Opening slash already gotten
	if self.token == token.QUOTIENT_ASSIGN {
		offset -= 1 // =
	}
	idx := self.idxOf(offset)

	pattern, err := self.scanString(offset)
	endOffset := self.chrOffset

	self.next()
	if err == nil {
		pattern = pattern[1 : len(pattern)-1]
	}

	flags := ""
	if self.token == token.IDENTIFIER { // gim

		flags = self.literal
		self.next()
		endOffset = self.chrOffset
	}

	// TODO 15.10
	{
		// Test during parsing that this is a valid regular expression
		// Sorry, (?=) and (?!) are invalid (for now)
		pattern := TransformRegExp(pattern)
		_, err := regexp.Compile(pattern)
		if err != nil {
			self.error(idx, "Invalid regular expression: %s", err.Error()[22:]) // Skip redundant "parse regexp error"
		}
	}

	literal := self.source[offset : endOffset-1]

	return &RegExpLiteral{
		Idx:     idx,
		Literal: literal,
		Pattern: pattern,
		Flags:   flags,
	}
}

func (self *_parser) parseVariableDeclaration() (string, Expression) {

	if self.token != token.IDENTIFIER {
		idx := self.expect(token.IDENTIFIER)
		self.nextStatement()
		return "", &BadExpression{From: idx, To: self.idx}
	}

	literal := self.literal
	idx := self.idx
	self.next()
	node := &VariableExpression{
		Name: literal,
		Idx:  idx,
	}

	if self.token == token.ASSIGN {
		self.next()
		node.Initializer = self.parseAssignmentExpression()
	}

	return literal, node
}

func (self *_parser) parseVariableDeclarationList() []Expression {

	var list []Expression

	for {
		name, definition := self.parseVariableDeclaration()
		list = append(list, definition)
		self.scope.addVariable(name)

		if self.token != token.COMMA {
			break
		}
		self.next()
	}

	return list
}

func (self *_parser) parseObjectPropertyKey() (string, string) {
	idx, tkn, literal := self.idx, self.token, self.literal
	value := ""
	self.next()
	switch tkn {
	case token.IDENTIFIER:
		value = literal
	case token.NUMBER:
		var err error
		_, err = parseNumberLiteral(literal)
		if err != nil {
			self.error(idx, err.Error())
		} else {
			value = literal
		}
	case token.STRING:
		value = parseStringLiteral(literal[1 : len(literal)-1])
	default:
		// null, false, class, etc.
		if matchIdentifier.MatchString(literal) {
			value = literal
		}
	}
	return literal, value
}

func (self *_parser) parseObjectProperty() Property {

	literal, value := self.parseObjectPropertyKey()
	if literal == "get" && self.token != token.COLON {
		idx := self.idx
		_, value := self.parseObjectPropertyKey()
		self.expect(token.LEFT_PARENTHESIS)
		self.expect(token.RIGHT_PARENTHESIS)

		node := &FunctionExpression{
			Function: idx,
		}
		self._parseFunction(node, "", false)
		return Property{
			Key:   value,
			Kind:  "get",
			Value: node,
		}
	} else if literal == "set" && self.token != token.COLON {
		idx := self.idx
		_, value := self.parseObjectPropertyKey()
		parameterList := self.parseFunctionParameterList()
		node := &FunctionExpression{
			Function:      idx,
			ParameterList: parameterList,
		}
		self._parseFunction(node, "", false)
		return Property{
			Key:   value,
			Kind:  "set",
			Value: node,
		}
	}

	self.expect(token.COLON)

	return Property{
		Key:   value,
		Kind:  "value",
		Value: self.parseAssignmentExpression(),
	}
}

func (self *_parser) parseObjectLiteral() Expression {
	var value []Property
	idx0 := self.expect(token.LEFT_BRACE)
	for self.token != token.RIGHT_BRACE && self.token != token.EOF {
		property := self.parseObjectProperty()
		value = append(value, property)
		if self.token == token.COMMA {
			self.next()
			continue
		}
	}
	idx1 := self.expect(token.RIGHT_BRACE)

	return &ObjectLiteral{
		LeftBrace:  idx0,
		RightBrace: idx1,
		Value:      value,
	}
}

func (self *_parser) parseArrayLiteral() Expression {

	idx0 := self.expect(token.LEFT_BRACKET)
	var value []Expression
	for self.token != token.RIGHT_BRACKET && self.token != token.EOF {
		if self.token == token.COMMA {
			self.next()
			value = append(value, nil)
			continue
		}
		value = append(value, self.parseAssignmentExpression())
		if self.token != token.RIGHT_BRACKET {
			self.expect(token.COMMA)
		}
	}
	idx1 := self.expect(token.RIGHT_BRACKET)

	return &ArrayLiteral{
		LeftBracket:  idx0,
		RightBracket: idx1,
		Value:        value,
	}
}

func (self *_parser) parseArgumentList() (argumentList []Expression, idx0, idx1 Idx) {
	idx0 = self.expect(token.LEFT_PARENTHESIS)
	if self.token != token.RIGHT_PARENTHESIS {
		for {
			argumentList = append(argumentList, self.parseAssignmentExpression())
			if self.token != token.COMMA {
				break
			}
			self.next()
		}
	}
	idx1 = self.expect(token.RIGHT_PARENTHESIS)
	return
}

func (self *_parser) parseCallExpression(left Expression) Expression {
	argumentList, idx0, idx1 := self.parseArgumentList()
	return &CallExpression{
		Callee:           left,
		LeftParenthesis:  idx0,
		ArgumentList:     argumentList,
		RightParenthesis: idx1,
	}
}

func (self *_parser) parseDotMember(left Expression) Expression {
	period := self.expect(token.PERIOD)

	literal := self.literal
	idx := self.idx

	if !matchIdentifier.MatchString(literal) {
		self.expect(token.IDENTIFIER)
		self.nextStatement()
		return &BadExpression{From: period, To: self.idx}
	}

	self.next()

	return &DotExpression{
		Left: left,
		Identifier: Identifier{
			Idx:  idx,
			Name: literal,
		},
	}
}

func (self *_parser) parseBracketMember(left Expression) Expression {
	idx0 := self.expect(token.LEFT_BRACKET)
	member := self.parseExpression()
	idx1 := self.expect(token.RIGHT_BRACKET)
	return &BracketExpression{
		LeftBracket:  idx0,
		Left:         left,
		Member:       member,
		RightBracket: idx1,
	}
}

func (self *_parser) parseNewExpression() Expression {
	idx := self.expect(token.NEW)
	callee := self.parseLeftHandSideExpression()
	node := &NewExpression{
		New:    idx,
		Callee: callee,
	}
	if self.token == token.LEFT_PARENTHESIS {
		argumentList, idx0, idx1 := self.parseArgumentList()
		node.ArgumentList = argumentList
		node.LeftParenthesis = idx0
		node.RightParenthesis = idx1
	}
	return node
}

func (self *_parser) parseLeftHandSideExpression() Expression {

	var left Expression
	if self.token == token.NEW {
		left = self.parseNewExpression()
	} else {
		left = self.parsePrimaryExpression()
	}

	for {
		if self.token == token.PERIOD {
			left = self.parseDotMember(left)
		} else if self.token == token.LEFT_BRACE {
			left = self.parseBracketMember(left)
		} else {
			break
		}
	}

	return left
}

func (self *_parser) parseLeftHandSideExpressionAllowCall() Expression {

	var left Expression
	if self.token == token.NEW {
		left = self.parseNewExpression()
	} else {
		left = self.parsePrimaryExpression()
	}

	for {
		if self.token == token.PERIOD {
			left = self.parseDotMember(left)
		} else if self.token == token.LEFT_BRACKET {
			left = self.parseBracketMember(left)
		} else if self.token == token.LEFT_PARENTHESIS {
			left = self.parseCallExpression(left)
		} else {
			break
		}
	}

	return left
}

func (self *_parser) parsePostfixExpression() Expression {
	operand := self.parseLeftHandSideExpressionAllowCall()

	switch self.token {
	case token.INCREMENT, token.DECREMENT:
		// Make sure there is no line terminator here
		if self.implicitSemicolon {
			break
		}
		tkn := self.token
		idx := self.idx
		self.next()
		switch operand.(type) {
		case *Identifier, *DotExpression, *BracketExpression:
		default:
			self.error(idx, "Invalid left-hand side in assignment")
			self.nextStatement()
			return &BadExpression{From: idx, To: self.idx}
		}
		return &UnaryExpression{
			Operator: tkn,
			Idx:      idx,
			Operand:  operand,
			Postfix:  true,
		}
	}

	return operand
}

func (self *_parser) parseUnaryExpression() Expression {

	switch self.token {
	case token.PLUS, token.MINUS, token.NOT, token.BITWISE_NOT:
		fallthrough
	case token.DELETE, token.VOID, token.TYPEOF:
		tkn := self.token
		idx := self.idx
		self.next()
		return &UnaryExpression{
			Operator: tkn,
			Idx:      idx,
			Operand:  self.parseUnaryExpression(),
		}
	case token.INCREMENT, token.DECREMENT:
		tkn := self.token
		idx := self.idx
		self.next()
		operand := self.parseUnaryExpression()
		switch operand.(type) {
		case *Identifier, *DotExpression, *BracketExpression:
		default:
			self.error(idx, "Invalid left-hand side in assignment")
			self.nextStatement()
			return &BadExpression{From: idx, To: self.idx}
		}
		return &UnaryExpression{
			Operator: tkn,
			Idx:      idx,
			Operand:  operand,
		}
	}

	return self.parsePostfixExpression()
}

func (self *_parser) parseMultiplicativeExpression() Expression {
	next := self.parseUnaryExpression
	left := next()

	for self.token == token.MULTIPLY || self.token == token.SLASH ||
		self.token == token.REMAINDER {
		tkn := self.token
		self.next()
		left = &BinaryExpression{
			Operator: tkn,
			Left:     left,
			Right:    next(),
		}
	}

	return left
}

func (self *_parser) parseAdditiveExpression() Expression {
	next := self.parseMultiplicativeExpression
	left := next()

	for self.token == token.PLUS || self.token == token.MINUS {
		tkn := self.token
		self.next()
		left = &BinaryExpression{
			Operator: tkn,
			Left:     left,
			Right:    next(),
		}
	}

	return left
}

func (self *_parser) parseShiftExpression() Expression {
	next := self.parseAdditiveExpression
	left := next()

	for self.token == token.SHIFT_LEFT || self.token == token.SHIFT_RIGHT ||
		self.token == token.UNSIGNED_SHIFT_RIGHT {
		tkn := self.token
		self.next()
		left = &BinaryExpression{
			Operator: tkn,
			Left:     left,
			Right:    next(),
		}
	}

	return left
}

func (self *_parser) parseRelationalExpression() Expression {
	next := self.parseShiftExpression
	left := next()

	allowIn := self.scope.allowIn
	self.scope.allowIn = true
	defer func() {
		self.scope.allowIn = allowIn
	}()

	switch self.token {
	case token.LESS, token.LESS_OR_EQUAL, token.GREATER, token.GREATER_OR_EQUAL:
		tkn := self.token
		self.next()
		return &BinaryExpression{
			Operator:   tkn,
			Left:       left,
			Right:      self.parseRelationalExpression(),
			Comparison: true,
		}
	case token.INSTANCEOF:
		tkn := self.token
		self.next()
		return &BinaryExpression{
			Operator: tkn,
			Left:     left,
			Right:    self.parseRelationalExpression(),
		}
	case token.IN:
		if !allowIn {
			return left
		}
		tkn := self.token
		self.next()
		return &BinaryExpression{
			Operator: tkn,
			Left:     left,
			Right:    self.parseRelationalExpression(),
		}
	}

	return left
}

func (self *_parser) parseEqualityExpression() Expression {
	next := self.parseRelationalExpression
	left := next()

	for self.token == token.EQUAL || self.token == token.NOT_EQUAL ||
		self.token == token.STRICT_EQUAL || self.token == token.STRICT_NOT_EQUAL {
		tkn := self.token
		self.next()
		left = &BinaryExpression{
			Operator:   tkn,
			Left:       left,
			Right:      next(),
			Comparison: true,
		}
	}

	return left
}

func (self *_parser) parseBitwiseAndExpression() Expression {
	next := self.parseEqualityExpression
	left := next()

	for self.token == token.AND {
		tkn := self.token
		self.next()
		left = &BinaryExpression{
			Operator: tkn,
			Left:     left,
			Right:    next(),
		}
	}

	return left
}

func (self *_parser) parseBitwiseExclusiveOrExpression() Expression {
	next := self.parseBitwiseAndExpression
	left := next()

	for self.token == token.EXCLUSIVE_OR {
		tkn := self.token
		self.next()
		left = &BinaryExpression{
			Operator: tkn,
			Left:     left,
			Right:    next(),
		}
	}

	return left
}

func (self *_parser) parseBitwiseOrExpression() Expression {
	next := self.parseBitwiseExclusiveOrExpression
	left := next()

	for self.token == token.OR {
		tkn := self.token
		self.next()
		left = &BinaryExpression{
			Operator: tkn,
			Left:     left,
			Right:    next(),
		}
	}

	return left
}

func (self *_parser) parseLogicalAndExpression() Expression {
	next := self.parseBitwiseOrExpression
	left := next()

	for self.token == token.LOGICAL_AND {
		tkn := self.token
		self.next()
		left = &BinaryExpression{
			Operator: tkn,
			Left:     left,
			Right:    next(),
		}
	}

	return left
}

func (self *_parser) parseLogicalOrExpression() Expression {
	next := self.parseLogicalAndExpression
	left := next()

	for self.token == token.LOGICAL_OR {
		tkn := self.token
		self.next()
		left = &BinaryExpression{
			Operator: tkn,
			Left:     left,
			Right:    next(),
		}
	}

	return left
}

func (self *_parser) parseConditionlExpression() Expression {
	left := self.parseLogicalOrExpression()

	if self.token == token.QUESTION_MARK {
		self.next()
		consequent := self.parseAssignmentExpression()
		self.expect(token.COLON)
		return &ConditionalExpression{
			Test:       left,
			Consequent: consequent,
			Alternate:  self.parseAssignmentExpression(),
		}
	}

	return left
}

func (self *_parser) parseAssignmentExpression() Expression {
	left := self.parseConditionlExpression()
	var operator token.Token
	switch self.token {
	case token.ASSIGN:
		operator = self.token
	case token.ADD_ASSIGN:
		operator = token.PLUS
	case token.SUBTRACT_ASSIGN:
		operator = token.MINUS
	case token.MULTIPLY_ASSIGN:
		operator = token.MULTIPLY
	case token.QUOTIENT_ASSIGN:
		operator = token.SLASH
	case token.REMAINDER_ASSIGN:
		operator = token.REMAINDER
	case token.AND_ASSIGN:
		operator = token.AND
	case token.AND_NOT_ASSIGN:
		operator = token.AND_NOT
	case token.OR_ASSIGN:
		operator = token.OR
	case token.EXCLUSIVE_OR_ASSIGN:
		operator = token.EXCLUSIVE_OR
	case token.SHIFT_LEFT_ASSIGN:
		operator = token.SHIFT_LEFT
	case token.SHIFT_RIGHT_ASSIGN:
		operator = token.SHIFT_RIGHT
	case token.UNSIGNED_SHIFT_RIGHT_ASSIGN:
		operator = token.UNSIGNED_SHIFT_RIGHT
	}

	if operator != 0 {
		idx := self.idx
		self.next()
		switch left.(type) {
		case *Identifier, *DotExpression, *BracketExpression:
		default:
			self.error(left.Idx0(), "Invalid left-hand side in assignment")
			self.nextStatement()
			return &BadExpression{From: idx, To: self.idx}
		}
		return &AssignExpression{
			Left:     left,
			Operator: operator,
			Right:    self.parseAssignmentExpression(),
		}
	}

	return left
}

func (self *_parser) parseExpression() Expression {
	next := self.parseAssignmentExpression
	left := next()

	if self.token == token.COMMA {
		sequence := []Expression{left}
		for {
			if self.token != token.COMMA {
				break
			}
			self.next()
			sequence = append(sequence, next())
		}
		return &SequenceExpression{
			Sequence: sequence,
		}
	}

	return left
}
