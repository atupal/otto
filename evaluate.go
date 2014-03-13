package otto

import (
	"fmt"
	"runtime"

	"github.com/robertkrimen/otto/parser"
	"github.com/robertkrimen/otto/parser/token"
)

func (self *_runtime) evaluateBody(body []parser.Statement) Value {
	bodyValue := Value{}
	for _, node := range body {
		value := self.evaluate(node)
		if value.isResult() {
			return value
		}
		if !value.isEmpty() {
			// We have GetValue here to (for example) trigger a
			// ReferenceError (of the not defined variety)
			// Not sure if this is the best way to error out early
			// for such errors or if there is a better way
			bodyValue = self.GetValue(value)
		}
	}
	return bodyValue
}

func (self *_runtime) evaluate(node parser.Node) Value {
	defer func() {
		// This defer is lame (unecessary overhead)
		// It would be better to mark the errors at the source
		if caught := recover(); caught != nil {
			switch caught := caught.(type) {
			case _error:
				if caught.Line == -1 {
					//caught.Line = node.position()
				}
				panic(caught) // Panic the modified _error
			}
			panic(caught)
		}
	}()

	// Allow interpreter interruption
	// If the Interrupt channel is nil, then
	// we avoid runtime.Gosched() overhead (if any)
	if self.Otto.Interrupt != nil {
		runtime.Gosched()
		select {
		case value := <-self.Otto.Interrupt:
			value()
		default:
		}
	}

	switch node := node.(type) {

	case *parser.VariableExpression:
		return self.evaluateVariableExpression(node)

	case *parser.VarStatement:
		// Variables are already defined, this is initialization only
		for _, variable := range node.List {
			self.evaluateVariableExpression(variable.(*parser.VariableExpression))
		}
		return Value{}

	case *parser.Program:
		self.functionDeclaration(node.FunctionList)
		self.variableDeclaration(node.VariableList)
		return self.evaluateBody(node.Body)

	case *parser.ExpressionStatement:
		return self.evaluate(node.Expression)

	case *parser.BlockStatement:
		return self.evaluateBlockStatement(node)

	case *parser.NullLiteral:
		return NullValue()

	case *parser.BooleanLiteral:
		return toValue_bool(node.Value)

	case *parser.StringLiteral:
		return toValue_string(node.Value)

	case *parser.NumberLiteral:
		return toValue_float64(stringToFloat(node.Literal))

	case *parser.ObjectLiteral:
		return self.evaluateObject(node)

	case *parser.RegExpLiteral:
		return self.evaluateRegExpLiteral(node)

	case *parser.ArrayLiteral:
		return self.evaluateArray(node)

	case *parser.Identifier:
		return self.evaluateIdentifier(node)

	case *parser.LabelledStatement:
		self.labels = append(self.labels, node.Label.Name)
		defer func() {
			if len(self.labels) > 0 {
				self.labels = self.labels[:len(self.labels)-1] // Pop the label
			} else {
				self.labels = nil
			}
		}()
		return self.evaluate(node.Statement)

	case *parser.BinaryExpression:
		if node.Comparison {
			return self.evaluateComparison(node)
		} else {
			return self.evaluateBinaryExpression(node)
		}

	case *parser.AssignExpression:
		return self.evaluateAssignExpression(node)

	case *parser.UnaryExpression:
		return self.evaluateUnaryExpression(node)

	case *parser.ReturnStatement:
		return self.evaluateReturnStatement(node)

	case *parser.IfStatement:
		return self.evaluateIfStatement(node)

	case *parser.DoWhileStatement:
		return self.evaluateDoWhileStatement(node)

	case *parser.WhileStatement:
		return self.evaluateWhileStatement(node)

	case *parser.CallExpression:
		return self.evaluateCall(node, nil)

	case *parser.BranchStatement:
		target := ""
		if node.Label != nil {
			target = node.Label.Name
		}
		switch node.Token {
		case token.BREAK:
			return toValue(newBreakResult(target))
		case token.CONTINUE:
			return toValue(newContinueResult(target))
		}

	case *parser.SwitchStatement:
		return self.evaluateSwitchStatement(node)

	case *parser.ForStatement:
		return self.evaluateForStatement(node)

	case *parser.ForInStatement:
		return self.evaluateForInStatement(node)

	case *parser.ThrowStatement:
		return self.evaluateThrowStatement(node)

	case *parser.EmptyStatement:
		return Value{}

	case *parser.TryStatement:
		return self.evaluateTryStatement(node)

	case *parser.DotExpression:
		return self.evaluateDotExpression(node)

	case *parser.BracketExpression:
		return self.evaluateBracketExpression(node)

	case *parser.NewExpression:
		return self.evaluateNew(node)

	case *parser.ConditionalExpression:
		return self.evaluateConditionalExpression(node)

	case *parser.ThisExpression:
		return toValue_object(self._executionContext(0).this)

	case *parser.SequenceExpression:
		return self.evaluateSequenceExpression(node)

	case *parser.WithStatement:
		return self.evaluateWithStatement(node)

	case *parser.FunctionExpression:
		return self.evaluateFunction(node)

	}

	panic(fmt.Sprintf("evaluate: Here be dragons: %T %v", node, node))
}
