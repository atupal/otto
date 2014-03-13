package parser

import (
	"github.com/robertkrimen/otto/parser/token"
)

type Node interface {
	Idx0() Idx
	Idx1() Idx
}

// ========== //
// Expression //
// ========== //

type (
	ArrayLiteral struct {
		LeftBracket  Idx
		RightBracket Idx
		Value        []Expression
	}

	AssignExpression struct {
		Operator token.Token
		Left     Expression
		Right    Expression
	}

	BadExpression struct {
		From Idx
		To   Idx
	}

	BinaryExpression struct {
		Operator   token.Token
		Left       Expression
		Right      Expression
		Comparison bool
	}

	BooleanLiteral struct {
		Idx     Idx
		Literal string
		Value   bool
	}

	BracketExpression struct {
		Left         Expression
		Member       Expression
		LeftBracket  Idx
		RightBracket Idx
	}

	CallExpression struct {
		Callee           Expression
		LeftParenthesis  Idx
		ArgumentList     []Expression
		RightParenthesis Idx
	}

	ConditionalExpression struct {
		Test       Expression
		Consequent Expression
		Alternate  Expression
	}

	DotExpression struct {
		Left       Expression
		Identifier Identifier
	}

	Expression interface {
		Node
		_expressionNode()
	}

	FunctionExpression struct {
		Function Idx
		Body     Statement
		Source   string

		ParameterList []string
		FunctionList  []Declaration
		VariableList  []Declaration
	}

	Identifier struct {
		Name string
		Idx  Idx
	}

	NewExpression struct {
		New              Idx
		Callee           Expression
		LeftParenthesis  Idx
		ArgumentList     []Expression
		RightParenthesis Idx
	}

	NullLiteral struct {
		Idx     Idx
		Literal string
	}

	NumberLiteral struct {
		Idx     Idx
		Literal string
		Value   interface{}
	}

	ObjectLiteral struct {
		LeftBrace  Idx
		RightBrace Idx
		Value      []Property
	}

	Property struct {
		Key   string
		Kind  string
		Value Expression
	}

	RegExpLiteral struct {
		Idx     Idx
		Literal string
		Pattern string
		Flags   string
	}

	SequenceExpression struct {
		Sequence []Expression
	}

	StringLiteral struct {
		Idx     Idx
		Literal string
		Value   string
	}

	ThisExpression struct {
		Idx
	}

	UnaryExpression struct {
		Operator token.Token
		Idx      Idx // If a prefix operation
		Operand  Expression
		Postfix  bool
	}

	VariableExpression struct {
		Name        string
		Idx         Idx
		Initializer Expression
	}
)

// _expressionNode

func (*ArrayLiteral) _expressionNode()          {}
func (*AssignExpression) _expressionNode()      {}
func (*BadExpression) _expressionNode()         {}
func (*BinaryExpression) _expressionNode()      {}
func (*BooleanLiteral) _expressionNode()        {}
func (*BracketExpression) _expressionNode()     {}
func (*CallExpression) _expressionNode()        {}
func (*ConditionalExpression) _expressionNode() {}
func (*DotExpression) _expressionNode()         {}
func (*FunctionExpression) _expressionNode()    {}
func (*Identifier) _expressionNode()            {}
func (*NewExpression) _expressionNode()         {}
func (*NullLiteral) _expressionNode()           {}
func (*NumberLiteral) _expressionNode()         {}
func (*ObjectLiteral) _expressionNode()         {}
func (*RegExpLiteral) _expressionNode()         {}
func (*SequenceExpression) _expressionNode()    {}
func (*StringLiteral) _expressionNode()         {}
func (*ThisExpression) _expressionNode()        {}
func (*UnaryExpression) _expressionNode()       {}
func (*VariableExpression) _expressionNode()    {}

// ========= //
// Statement //
// ========= //

type (
	Statement interface {
		Node
		_statementNode()
	}

	BadStatement struct {
		From Idx
		To   Idx
	}

	BlockStatement struct {
		LeftBrace  Idx
		List       []Statement
		RightBrace Idx
	}

	BranchStatement struct {
		Idx   Idx
		Token token.Token
		Label *Identifier
	}

	CaseStatement struct {
		Case       Idx
		Test       Expression
		Consequent []Statement
	}

	CatchStatement struct {
		Catch     Idx
		Parameter *Identifier
		Body      Statement
	}

	DoWhileStatement struct {
		Do   Idx
		Test Expression
		Body Statement
	}

	EmptyStatement struct {
		Semicolon Idx
	}

	ExpressionStatement struct {
		Expression Expression
	}

	ForInStatement struct {
		For    Idx
		Into   Expression
		Source Expression
		Body   Statement
	}

	ForStatement struct {
		For         Idx
		Initializer Expression
		Update      Expression
		Test        Expression
		Body        Statement
	}

	IfStatement struct {
		If         Idx
		Test       Expression
		Consequent Statement
		Alternate  Statement
	}

	LabelledStatement struct {
		Label     *Identifier
		Colon     Idx
		Statement Statement
	}

	ReturnStatement struct {
		Return   Idx
		Argument Expression
	}

	SwitchStatement struct {
		Switch       Idx
		Discriminant Expression
		Default      int
		Body         []*CaseStatement
	}

	ThrowStatement struct {
		Throw    Idx
		Argument Expression
	}

	TryStatement struct {
		Try     Idx
		Body    Statement
		Catch   *CatchStatement
		Finally Statement
	}

	VarStatement struct {
		Var  Idx
		List []Expression
	}

	WhileStatement struct {
		While Idx
		Test  Expression
		Body  Statement
	}

	WithStatement struct {
		With   Idx
		Object Expression
		Body   Statement
	}
)

// _statementNode

func (*BadStatement) _statementNode()        {}
func (*BlockStatement) _statementNode()      {}
func (*BranchStatement) _statementNode()     {}
func (*CaseStatement) _statementNode()       {}
func (*CatchStatement) _statementNode()      {}
func (*DoWhileStatement) _statementNode()    {}
func (*EmptyStatement) _statementNode()      {}
func (*ExpressionStatement) _statementNode() {}
func (*ForInStatement) _statementNode()      {}
func (*ForStatement) _statementNode()        {}
func (*IfStatement) _statementNode()         {}
func (*LabelledStatement) _statementNode()   {}
func (*ReturnStatement) _statementNode()     {}
func (*SwitchStatement) _statementNode()     {}
func (*ThrowStatement) _statementNode()      {}
func (*TryStatement) _statementNode()        {}
func (*VarStatement) _statementNode()        {}
func (*WhileStatement) _statementNode()      {}
func (*WithStatement) _statementNode()       {}

// ==== //
// Node //
// ==== //

type Program struct {
	Body []Statement

	FunctionList []Declaration
	VariableList []Declaration
}

// ==== //
// Idx0 //
// ==== //

func (self *ArrayLiteral) Idx0() Idx          { return self.LeftBracket }
func (self *AssignExpression) Idx0() Idx      { return self.Left.Idx0() }
func (self *BadExpression) Idx0() Idx         { return self.From }
func (self *BinaryExpression) Idx0() Idx      { return self.Left.Idx0() }
func (self *BooleanLiteral) Idx0() Idx        { return self.Idx }
func (self *BracketExpression) Idx0() Idx     { return self.Left.Idx0() }
func (self *CallExpression) Idx0() Idx        { return self.Callee.Idx0() }
func (self *ConditionalExpression) Idx0() Idx { return self.Test.Idx0() }
func (self *DotExpression) Idx0() Idx         { return self.Left.Idx0() }
func (self *FunctionExpression) Idx0() Idx    { return self.Function }
func (self *Identifier) Idx0() Idx            { return self.Idx }
func (self *NewExpression) Idx0() Idx         { return self.New }
func (self *NullLiteral) Idx0() Idx           { return self.Idx }
func (self *NumberLiteral) Idx0() Idx         { return self.Idx }
func (self *ObjectLiteral) Idx0() Idx         { return self.LeftBrace }
func (self *RegExpLiteral) Idx0() Idx         { return self.Idx }
func (self *SequenceExpression) Idx0() Idx    { return self.Sequence[0].Idx0() }
func (self *StringLiteral) Idx0() Idx         { return self.Idx }
func (self *ThisExpression) Idx0() Idx        { return self.Idx }
func (self *UnaryExpression) Idx0() Idx       { return self.Idx }
func (self *VariableExpression) Idx0() Idx    { return self.Idx }

func (self *BadStatement) Idx0() Idx        { return self.From }
func (self *BlockStatement) Idx0() Idx      { return self.LeftBrace }
func (self *BranchStatement) Idx0() Idx     { return self.Idx }
func (self *CaseStatement) Idx0() Idx       { return self.Case }
func (self *CatchStatement) Idx0() Idx      { return self.Catch }
func (self *DoWhileStatement) Idx0() Idx    { return self.Do }
func (self *EmptyStatement) Idx0() Idx      { return self.Semicolon }
func (self *ExpressionStatement) Idx0() Idx { return self.Expression.Idx0() }
func (self *ForInStatement) Idx0() Idx      { return self.For }
func (self *ForStatement) Idx0() Idx        { return self.For }
func (self *IfStatement) Idx0() Idx         { return self.If }
func (self *LabelledStatement) Idx0() Idx   { return self.Label.Idx0() }
func (self *Program) Idx0() Idx             { return self.Body[0].Idx0() }
func (self *ReturnStatement) Idx0() Idx     { return self.Return }
func (self *SwitchStatement) Idx0() Idx     { return self.Switch }
func (self *ThrowStatement) Idx0() Idx      { return self.Throw }
func (self *TryStatement) Idx0() Idx        { return self.Try }
func (self *VarStatement) Idx0() Idx        { return self.Var }
func (self *WhileStatement) Idx0() Idx      { return self.While }
func (self *WithStatement) Idx0() Idx       { return self.With }

// ==== //
// Idx1 //
// ==== //

func (self *ArrayLiteral) Idx1() Idx          { return self.RightBracket }
func (self *AssignExpression) Idx1() Idx      { return self.Right.Idx1() }
func (self *BadExpression) Idx1() Idx         { return self.To }
func (self *BinaryExpression) Idx1() Idx      { return self.Right.Idx1() }
func (self *BooleanLiteral) Idx1() Idx        { return Idx(int(self.Idx) + len(self.Literal)) }
func (self *BracketExpression) Idx1() Idx     { return self.RightBracket + 1 }
func (self *CallExpression) Idx1() Idx        { return self.RightParenthesis + 1 }
func (self *ConditionalExpression) Idx1() Idx { return self.Test.Idx1() }
func (self *DotExpression) Idx1() Idx         { return self.Identifier.Idx1() }
func (self *FunctionExpression) Idx1() Idx    { return self.Body.Idx1() }
func (self *Identifier) Idx1() Idx            { return Idx(int(self.Idx) + len(self.Name)) }
func (self *NewExpression) Idx1() Idx         { return self.RightParenthesis + 1 }
func (self *NullLiteral) Idx1() Idx           { return Idx(int(self.Idx) + 4) } // "null"
func (self *NumberLiteral) Idx1() Idx         { return Idx(int(self.Idx) + len(self.Literal)) }
func (self *ObjectLiteral) Idx1() Idx         { return self.RightBrace }
func (self *RegExpLiteral) Idx1() Idx         { return Idx(int(self.Idx) + len(self.Literal)) }
func (self *SequenceExpression) Idx1() Idx    { return self.Sequence[0].Idx1() }
func (self *StringLiteral) Idx1() Idx         { return Idx(int(self.Idx) + len(self.Literal)) }
func (self *ThisExpression) Idx1() Idx        { return self.Idx }
func (self *UnaryExpression) Idx1() Idx {
	if self.Postfix {
		return self.Operand.Idx1() + 2 // ++ --
	}
	return self.Operand.Idx1()
}
func (self *VariableExpression) Idx1() Idx {
	if self.Initializer == nil {
		return Idx(int(self.Idx) + len(self.Name) + 1)
	}
	return self.Initializer.Idx1()
}

func (self *BadStatement) Idx1() Idx        { return self.To }
func (self *BlockStatement) Idx1() Idx      { return self.RightBrace + 1 }
func (self *BranchStatement) Idx1() Idx     { return self.Idx }
func (self *CaseStatement) Idx1() Idx       { return self.Consequent[len(self.Consequent)-1].Idx1() }
func (self *CatchStatement) Idx1() Idx      { return self.Body.Idx1() }
func (self *DoWhileStatement) Idx1() Idx    { return self.Test.Idx1() }
func (self *EmptyStatement) Idx1() Idx      { return self.Semicolon + 1 }
func (self *ExpressionStatement) Idx1() Idx { return self.Expression.Idx1() }
func (self *ForInStatement) Idx1() Idx      { return self.Body.Idx1() }
func (self *ForStatement) Idx1() Idx        { return self.Body.Idx1() }
func (self *IfStatement) Idx1() Idx {
	if self.Alternate != nil {
		return self.Alternate.Idx1()
	}
	return self.Consequent.Idx1()
}
func (self *LabelledStatement) Idx1() Idx { return self.Colon + 1 }
func (self *Program) Idx1() Idx           { return self.Body[len(self.Body)-1].Idx1() }
func (self *ReturnStatement) Idx1() Idx   { return self.Return }
func (self *SwitchStatement) Idx1() Idx   { return self.Body[len(self.Body)-1].Idx1() }
func (self *ThrowStatement) Idx1() Idx    { return self.Throw }
func (self *TryStatement) Idx1() Idx      { return self.Try }
func (self *VarStatement) Idx1() Idx      { return self.List[len(self.List)-1].Idx1() }
func (self *WhileStatement) Idx1() Idx    { return self.Body.Idx1() }
func (self *WithStatement) Idx1() Idx     { return self.Body.Idx1() }
