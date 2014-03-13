package parser

import (
	"github.com/robertkrimen/otto/parser/token"
)

func (self *_parser) parseBlockStatement() *BlockStatement {
	node := &BlockStatement{}
	node.LeftBrace = self.expect(token.LEFT_BRACE)
	node.List = self.parseStatementList()
	node.RightBrace = self.expect(token.RIGHT_BRACE)

	return node
}

func (self *_parser) parseEmptyStatement() Statement {
	idx := self.expect(token.SEMICOLON)
	return &EmptyStatement{Semicolon: idx}
}

func (self *_parser) parseStatementList() (list []Statement) {
	for self.token != token.RIGHT_BRACE && self.token != token.EOF {
		list = append(list, self.parseStatement())
	}

	return
}

func (self *_parser) parseStatement() Statement {

	if self.token == token.EOF {
		self.errorUnexpectedToken(self.token)
		return &BadStatement{From: self.idx, To: self.idx + 1}
	}

	switch self.token {
	case token.SEMICOLON:
		return self.parseEmptyStatement()
	case token.LEFT_BRACE:
		return self.parseBlockStatement()
	case token.IF:
		return self.parseIfStatement()
	case token.DO:
		return self.parseDoWhileStatement()
	case token.WHILE:
		return self.parseWhileStatement()
	case token.FOR:
		return self.parseForOrForInStatement()
	case token.BREAK:
		return self.parseBreakStatement()
	case token.CONTINUE:
		return self.parseContinueStatement()
	case token.WITH:
		return self.parseWithStatement()
	case token.VAR:
		return self.parseVariableStatement()
	case token.FUNCTION:
		self.parseFunction(true)
		return &EmptyStatement{}
	case token.SWITCH:
		return self.parseSwitchStatement()
	case token.RETURN:
		return self.parseReturnStatement()
	case token.THROW:
		return self.parseThrowStatement()
	case token.TRY:
		return self.parseTryStatement()
	}

	expression := self.parseExpression()

	if identifier, isIdentifier := expression.(*Identifier); isIdentifier && self.token == token.COLON {
		// LabelledStatement
		colon := self.idx
		self.next() // :
		label := identifier.Name
		for _, value := range self.scope.labels {
			if label == value {
				self.error(identifier.Idx0(), "Label '%s' already exists", label)
			}
		}
		self.scope.labels = append(self.scope.labels, label) // Push the label
		statement := self.parseStatement()
		self.scope.labels = self.scope.labels[:len(self.scope.labels)-1] // Pop the label
		return &LabelledStatement{
			Label:     identifier,
			Colon:     colon,
			Statement: statement,
		}
	}

	self.optionalSemicolon()

	return &ExpressionStatement{
		Expression: expression,
	}
}

func (self *_parser) parseTryStatement() Statement {

	node := &TryStatement{
		Try:  self.expect(token.TRY),
		Body: self.parseBlockStatement(),
	}

	if self.token == token.CATCH {
		catch := self.idx
		self.next()
		self.expect(token.LEFT_PARENTHESIS)
		if self.token != token.IDENTIFIER {
			self.expect(token.IDENTIFIER)
			self.nextStatement()
			return &BadStatement{From: catch, To: self.idx}
		} else {
			identifier := self.parseIdentifier()
			self.expect(token.RIGHT_PARENTHESIS)
			node.Catch = &CatchStatement{
				Catch:     catch,
				Parameter: &identifier,
				Body:      self.parseBlockStatement(),
			}
		}
	}

	if self.token == token.FINALLY {
		self.next()
		node.Finally = self.parseBlockStatement()
		self.semicolon()
	}

	if node.Catch == nil && node.Finally == nil {
		self.error(node.Try, "Missing catch or finally after try")
		return &BadStatement{From: node.Try, To: node.Body.Idx1()}
	}

	return node
}

func (self *_parser) parseFunctionParameterList() (list []string) {
	self.expect(token.LEFT_PARENTHESIS)
	for self.token != token.RIGHT_PARENTHESIS && self.token != token.EOF {
		if self.token != token.IDENTIFIER {
			self.expect(token.IDENTIFIER)
		}
		list = append(list, self.literal)
		self.next()
		if self.token != token.RIGHT_PARENTHESIS {
			self.expect(token.COMMA)
		}
	}
	self.expect(token.RIGHT_PARENTHESIS)
	return
}

func (self *_parser) parseFunction(declaration bool) Expression {

	node := &FunctionExpression{
		Function: self.expect(token.FUNCTION),
	}

	name := ""
	if self.token == token.IDENTIFIER {
		name = self.literal
		self.next()
		if declaration {
			self.scope.addFunction(name, node)
		}
	} else if declaration {
		// Use expect error handling
		self.expect(token.IDENTIFIER)
	}

	node.ParameterList = self.parseFunctionParameterList()
	self._parseFunction(node, name, declaration)

	node.Source = self.slice(node.Idx0(), node.Idx1())

	return node
}

func (self *_parser) _parseFunction(node *FunctionExpression, name string, declaration bool) {
	{
		self.openScope()
		if !declaration && name != "" {
			self.scope.addFunction(name, node)
		}
		inFunction := self.scope.inFunction
		self.scope.inFunction = true
		defer func() {
			self.scope.inFunction = inFunction
			self.closeScope()
		}()
		node.Body = self.parseBlockStatement()
		node.VariableList = self.scope.variableList
		node.FunctionList = self.scope.functionList
	}
}

func (self *_parser) parseReturnStatement() Statement {
	idx := self.expect(token.RETURN)

	if !self.scope.inFunction {
		self.error(idx, "Illegal return statement")
		self.nextStatement()
		return &BadStatement{From: idx, To: self.idx}
	}

	node := &ReturnStatement{
		Return: idx,
	}

	if !self.implicitSemicolon && self.token != token.SEMICOLON && self.token != token.RIGHT_BRACE && self.token != token.EOF {
		node.Argument = self.parseExpression()
	}

	self.semicolon()

	return node
}

func (self *_parser) parseThrowStatement() Statement {
	idx := self.expect(token.THROW)

	if self.implicitSemicolon {
		if self.chr == -1 { // Hackish
			self.error(idx, "Unexpected end of input")
		} else {
			self.error(idx, "Illegal newline after throw")
		}
		self.nextStatement()
		return &BadStatement{From: idx, To: self.idx}
	}

	node := &ThrowStatement{
		Argument: self.parseExpression(),
	}

	self.semicolon()

	return node
}

func (self *_parser) parseSwitchStatement() Statement {
	self.expect(token.SWITCH)
	self.expect(token.LEFT_PARENTHESIS)
	node := &SwitchStatement{
		Discriminant: self.parseExpression(),
		Default:      -1,
	}
	self.expect(token.RIGHT_PARENTHESIS)

	self.expect(token.LEFT_BRACE)

	inSwitch := self.scope.inSwitch
	self.scope.inSwitch = true
	defer func() {
		self.scope.inSwitch = inSwitch
	}()

	for index := 0; self.token != token.EOF; index++ {
		if self.token == token.RIGHT_BRACE {
			self.next()
			break
		}

		clause := self.parseCaseStatement()
		if clause.Test == nil {
			if node.Default != -1 {
				self.error(clause.Case, "Already saw a default in switch")
			}
			node.Default = index
		}
		node.Body = append(node.Body, clause)
	}

	return node
}

func (self *_parser) parseWithStatement() Statement {
	self.expect(token.WITH)
	self.expect(token.LEFT_PARENTHESIS)
	node := &WithStatement{
		Object: self.parseExpression(),
	}
	self.expect(token.RIGHT_PARENTHESIS)

	node.Body = self.parseStatement()

	return node
}

func (self *_parser) parseCaseStatement() *CaseStatement {

	node := &CaseStatement{
		Case: self.idx,
	}
	if self.token == token.DEFAULT {
		self.next()
	} else {
		self.expect(token.CASE)
		node.Test = self.parseExpression()
	}
	self.expect(token.COLON)

	for {
		if self.token == token.EOF ||
			self.token == token.RIGHT_BRACE ||
			self.token == token.CASE ||
			self.token == token.DEFAULT {
			break
		}
		node.Consequent = append(node.Consequent, self.parseStatement())

	}

	return node
}

func (self *_parser) parseIterationStatement() Statement {
	inIteration := self.scope.inIteration
	self.scope.inIteration = true
	defer func() {
		self.scope.inIteration = inIteration
	}()
	return self.parseStatement()
}

func (self *_parser) parseForIn(into Expression) *ForInStatement {

	// Already have consumed "<into> in"

	source := self.parseExpression()
	self.expect(token.RIGHT_PARENTHESIS)

	return &ForInStatement{
		Into:   into,
		Source: source,
		Body:   self.parseIterationStatement(),
	}
}

func (self *_parser) parseFor(initializer Expression) *ForStatement {

	// Already have consumed "<initializer> ;"

	var test, update Expression

	if self.token != token.SEMICOLON {
		test = self.parseExpression()
	}
	self.expect(token.SEMICOLON)

	if self.token != token.RIGHT_PARENTHESIS {
		update = self.parseExpression()
	}
	self.expect(token.RIGHT_PARENTHESIS)

	return &ForStatement{
		Initializer: initializer,
		Test:        test,
		Update:      update,
		Body:        self.parseIterationStatement(),
	}
}

func (self *_parser) parseForOrForInStatement() Statement {
	idx := self.expect(token.FOR)
	self.expect(token.LEFT_PARENTHESIS)

	var left []Expression

	isIn := false
	if self.token != token.SEMICOLON {

		allowIn := self.scope.allowIn
		self.scope.allowIn = false
		if self.token == token.VAR {

			self.next()
			// FIXME
			list := self.parseVariableDeclarationList()
			if len(list) == 1 && self.token == token.IN {
				self.next() // in
				isIn = true
				// We only want (there should be only) one _declaration
				// (12.2 Variable Statement)
				// FIXME
				left = append(left, list[0])
			} else {
				left = list
			}
		} else {
			left = append(left, self.parseExpression())
			if self.token == token.IN {
				self.next()
				isIn = true
			}
		}
		self.scope.allowIn = allowIn
	}

	if !isIn {
		self.expect(token.SEMICOLON)
		// FIXME
		return self.parseFor(&SequenceExpression{Sequence: left})
	} else {
		switch left[0].(type) {
		case *Identifier, *DotExpression, *BracketExpression, *VariableExpression:
		default:
			self.error(idx, "Invalid left-hand side in for-in")
			self.nextStatement()
			return &BadStatement{From: idx, To: self.idx}
		}
	}

	return self.parseForIn(left[0])
}

func (self *_parser) parseVariableStatement() *VarStatement {

	idx := self.expect(token.VAR)

	list := self.parseVariableDeclarationList()
	for _, variable := range list {
		switch variable := variable.(type) {
		case *VariableExpression:
			self.scope.addVariable(variable.Name)
		}
	}

	self.semicolon()

	return &VarStatement{
		Var:  idx,
		List: list,
	}
}

func (self *_parser) parseDoWhileStatement() Statement {
	inIteration := self.scope.inIteration
	self.scope.inIteration = true
	defer func() {
		self.scope.inIteration = inIteration
	}()

	self.expect(token.DO)
	node := &DoWhileStatement{}
	if self.token == token.LEFT_BRACE {
		node.Body = self.parseBlockStatement()
	} else {
		node.Body = self.parseStatement()
	}

	self.expect(token.WHILE)
	self.expect(token.LEFT_PARENTHESIS)
	node.Test = self.parseExpression()
	self.expect(token.RIGHT_PARENTHESIS)

	return node
}

func (self *_parser) parseWhileStatement() Statement {
	self.expect(token.WHILE)
	self.expect(token.LEFT_PARENTHESIS)
	node := &WhileStatement{
		Test: self.parseExpression(),
	}
	self.expect(token.RIGHT_PARENTHESIS)
	node.Body = self.parseIterationStatement()

	return node
}

func (self *_parser) parseIfStatement() Statement {
	self.expect(token.IF)
	self.expect(token.LEFT_PARENTHESIS)
	node := &IfStatement{
		Test: self.parseExpression(),
	}
	self.expect(token.RIGHT_PARENTHESIS)

	if self.token == token.LEFT_BRACE {
		node.Consequent = self.parseBlockStatement()
	} else {
		node.Consequent = self.parseStatement()
	}

	if self.token == token.ELSE {
		self.next()
		node.Alternate = self.parseStatement()
	}

	return node
}

func (self *_parser) parseSourceElement() Statement {
	tkn := self.token

	switch tkn {
	case token.CONST, token.LET, token.FUNCTION:
	default:
	case token.EOF:
		break
	}

	return self.parseStatement()
}

func (self *_parser) parseSourceElements() []Statement {
	body := []Statement(nil)

	for {
		if self.token != token.STRING {
			break
		}

		body = append(body, self.parseSourceElement())
	}

	for self.token != token.EOF {
		body = append(body, self.parseSourceElement())
	}

	return body
}

func (self *_parser) parseProgram() *Program {
	self.openScope()
	defer self.closeScope()
	return &Program{
		Body:         self.parseSourceElements(),
		VariableList: self.scope.variableList,
		FunctionList: self.scope.functionList,
	}
}

func (self *_parser) parseBreakStatement() Statement {
	idx := self.expect(token.BREAK)
	semicolon := self.implicitSemicolon
	if self.token == token.SEMICOLON {
		semicolon = true
		self.next()
	}

	if semicolon {
		self.implicitSemicolon = false
		if !self.scope.inIteration && !self.scope.inSwitch {
			goto illegal
		}
		return &BranchStatement{
			Idx:   idx,
			Token: token.BREAK,
		}
	}

	if self.token == token.IDENTIFIER {
		identifier := self.parseIdentifier()
		if !self.scope.hasLabel(identifier.Name) {
			self.error(idx, "Undefined label '%s'", identifier.Name)
			return &BadStatement{From: idx, To: identifier.Idx1()}
		}
		self.semicolon()
		return &BranchStatement{
			Idx:   idx,
			Token: token.BREAK,
			Label: &identifier,
		}
	}

	self.expect(token.IDENTIFIER)

illegal:
	self.error(idx, "Illegal break statement")
	self.nextStatement()
	return &BadStatement{From: idx, To: self.idx}
}

func (self *_parser) parseContinueStatement() Statement {
	idx := self.expect(token.CONTINUE)
	semicolon := self.implicitSemicolon
	if self.token == token.SEMICOLON {
		semicolon = true
		self.next()
	}

	if semicolon {
		self.implicitSemicolon = false
		if !self.scope.inIteration {
			goto illegal
		}
		return &BranchStatement{
			Idx:   idx,
			Token: token.CONTINUE,
		}
	}

	if self.token == token.IDENTIFIER {
		identifier := self.parseIdentifier()
		if !self.scope.hasLabel(identifier.Name) {
			self.error(idx, "Undefined label '%s'", identifier.Name)
			return &BadStatement{From: idx, To: identifier.Idx1()}
		}
		if !self.scope.inIteration {
			goto illegal
		}
		self.semicolon()
		return &BranchStatement{
			Idx:   idx,
			Token: token.CONTINUE,
			Label: &identifier,
		}
	}

	self.expect(token.IDENTIFIER)

illegal:
	self.error(idx, "Illegal continue statement")
	self.nextStatement()
	return &BadStatement{From: idx, To: self.idx}
}

// Find the next statement after an error (recover)
func (self *_parser) nextStatement() {
	for {
		switch self.token {
		case token.BREAK, token.CONST, token.CONTINUE,
			token.FOR, token.IF, token.RETURN, token.SWITCH,
			token.VAR, token.DO, token.TRY, token.LET, token.WITH,
			token.WHILE, token.THROW, token.CATCH, token.FINALLY:
			// Return only if parser made some progress since last
			// sync or if it has not reached 10 next calls without
			// progress. Otherwise consume at least one token to
			// avoid an endless parser loop
			if self.idx == self.recover.idx && self.recover.count < 10 {
				self.recover.count++
				return
			}
			if self.idx > self.recover.idx {
				self.recover.idx = self.idx
				self.recover.count = 0
				return
			}
			// Reaching here indicates a parser bug, likely an
			// incorrect token list in this function, but it only
			// leads to skipping of possibly correct code if a
			// previous error is present, and thus is preferred
			// over a non-terminating parse.
		case token.EOF:
			return
		}
		self.next()
	}
}
