package otto

import (
	"github.com/robertkrimen/otto/parser"
	"github.com/robertkrimen/otto/parser/token"
)

func (self *_runtime) evaluateTryStatement(node *parser.TryStatement) Value {
	tryCatchValue, exception := self.tryCatchEvaluate(func() Value {
		return self.evaluate(node.Body)
	})

	if exception && node.Catch != nil {

		lexicalEnvironment := self._executionContext(0).newDeclarativeEnvironment(self)
		defer func() {
			self._executionContext(0).LexicalEnvironment = lexicalEnvironment
		}()
		// TODO If necessary, convert TypeError<runtime> => TypeError
		// That, is, such errors can be thrown despite not being JavaScript "native"
		self.localSet(node.Catch.Parameter.Name, tryCatchValue)

		tryCatchValue, exception = self.tryCatchEvaluate(func() Value {
			return self.evaluate(node.Catch.Body)
		})
	}

	if node.Finally != nil {
		finallyValue := self.evaluate(node.Finally)
		if finallyValue.isResult() {
			return finallyValue
		}
	}

	if exception {
		panic(newException(tryCatchValue))
	}

	return tryCatchValue
}

//func (self *_runtime) evaluateVariableDeclarationList(node *_variableDeclarationListNode) Value {
//    for _, node := range node.VariableList {
//        self.evaluateVariableDeclaration(node)
//    }
//    return emptyValue()
//}

func (self *_runtime) evaluateThrowStatement(node *parser.ThrowStatement) Value {
	value := self.GetValue(self.evaluate(node.Argument))
	panic(newException(value))
}

func (self *_runtime) evaluateReturnStatement(node *parser.ReturnStatement) Value {
	value := UndefinedValue()
	if node.Argument != nil {
		value = self.GetValue(self.evaluate(node.Argument))
	}

	return toValue(newReturnResult(value))
}

func (self *_runtime) evaluateIfStatement(node *parser.IfStatement) Value {
	test := self.evaluate(node.Test)
	testValue := self.GetValue(test)
	if toBoolean(testValue) {
		return self.evaluate(node.Consequent)
	} else if node.Alternate != nil {
		return self.evaluate(node.Alternate)
	}

	return Value{}
}

func (self *_runtime) evaluateWithStatement(node *parser.WithStatement) Value {
	object := self.evaluate(node.Object)
	objectValue := self.GetValue(object)
	previousLexicalEnvironment, lexicalEnvironment := self._executionContext(0).newLexicalEnvironment(self.toObject(objectValue))
	lexicalEnvironment.ProvideThis = true
	defer func() {
		self._executionContext(0).LexicalEnvironment = previousLexicalEnvironment
	}()

	return self.evaluate(node.Body)
}

func (self *_runtime) evaluateBlockStatement(node *parser.BlockStatement) Value {

	//labelSet := node.labelSet

	value := self.evaluateBody(node.List)
	return value
	//if blockValue.evaluateBreak(labelSet) == resultBreak {
	//    return Value{}
	//}
	//return value
}

func (self *_runtime) evaluateDoWhileStatement(node *parser.DoWhileStatement) Value {

	labels := append(self.labels, "")
	self.labels = nil

	test := node.Test
	var body []parser.Statement
	switch tmp := node.Body.(type) {
	case *parser.BlockStatement:
		body = tmp.List
	default:
		body = append(body, node.Body)
	}

	result := Value{}
resultBreak:
	for {
		for _, node := range body {
			value := self.evaluate(node)
			switch value._valueType {
			case valueResult:
				switch value.evaluateBreakContinue(labels) {
				case resultReturn:
					return value
				case resultBreak:
					break resultBreak
				case resultContinue:
					goto resultContinue
				}
			case valueEmpty:
			default:
				result = value
			}
		}
	resultContinue:
		if !self.GetValue(self.evaluate(test)).isTrue() {
			// Stahp: do ... while (false)
			break
		}
	}
	return result
}

func (self *_runtime) evaluateWhileStatement(node *parser.WhileStatement) Value {

	test := node.Test
	labels := append(self.labels, "")
	self.labels = nil

	var body []parser.Statement
	switch tmp := node.Body.(type) {
	case *parser.BlockStatement:
		body = tmp.List
	default:
		body = append(body, node.Body)
	}

	result := Value{}
resultBreakContinue:
	for {
		if !self.GetValue(self.evaluate(test)).isTrue() {
			// Stahp: while (false) ...
			break
		}
		for _, node := range body {
			value := self.evaluate(node)
			switch value._valueType {
			case valueResult:
				switch value.evaluateBreakContinue(labels) {
				case resultReturn:
					return value
				case resultBreak:
					break resultBreakContinue
				case resultContinue:
					continue resultBreakContinue
				}
			case valueEmpty:
			default:
				result = value
			}
		}
	}
	return result
}

func (self *_runtime) evaluateForStatement(node *parser.ForStatement) Value {

	labels := append(self.labels, "")
	self.labels = nil

	initializer := node.Initializer
	test := node.Test
	update := node.Update

	if initializer != nil {
		initialResult := self.evaluate(initializer)
		self.GetValue(initialResult) // Side-effect trigger
	}

	var body []parser.Statement
	switch tmp := node.Body.(type) {
	case *parser.BlockStatement:
		body = tmp.List
	default:
		body = append(body, node.Body)
	}

	result := Value{}
resultBreak:
	for {
		if test != nil {
			testResult := self.evaluate(test)
			testResultValue := self.GetValue(testResult)
			if toBoolean(testResultValue) == false {
				break
			}
		}
		for _, node := range body {
			value := self.evaluate(node)
			switch value._valueType {
			case valueResult:
				switch value.evaluateBreakContinue(labels) {
				case resultReturn:
					return value
				case resultBreak:
					break resultBreak
				case resultContinue:
					goto resultContinue
				}
			case valueEmpty:
			default:
				result = value
			}
		}
	resultContinue:
		if update != nil {
			updateResult := self.evaluate(update)
			self.GetValue(updateResult) // Side-effect trigger
		}
	}
	return result
}

func (self *_runtime) evaluateForInStatement(node *parser.ForInStatement) Value {

	labels := append(self.labels, "")
	self.labels = nil

	source := self.evaluate(node.Source)
	sourceValue := self.GetValue(source)

	switch sourceValue._valueType {
	case valueUndefined, valueNull:
		return emptyValue()
	}

	sourceObject := self.toObject(sourceValue)

	into := node.Into

	var body []parser.Statement
	switch tmp := node.Body.(type) {
	case *parser.BlockStatement:
		body = tmp.List
	default:
		body = append(body, node.Body)
	}

	result := Value{}
	object := sourceObject
	for object != nil {
		enumerateValue := Value{}
		object.enumerate(false, func(name string) bool {
			into := self.evaluate(into)
			// In the case of: for (var abc in def) ...
			if into.reference() == nil {
				identifier := toString(into)
				// TODO Should be true or false (strictness) depending on context
				into = toValue(getIdentifierReference(self.LexicalEnvironment(), identifier, false, node))
			}
			self.PutValue(into.reference(), toValue_string(name))
			for _, node := range body {
				value := self.evaluate(node)
				switch value._valueType {
				case valueResult:
					switch value.evaluateBreakContinue(labels) {
					case resultReturn:
						enumerateValue = value
						return false
					case resultBreak:
						object = nil
						return false
					case resultContinue:
						return true
					}
				case valueEmpty:
				default:
					enumerateValue = value
				}
			}
			return true
		})
		if object == nil {
			break
		}
		object = object.prototype
		if !enumerateValue.isEmpty() {
			result = enumerateValue
		}
	}
	return result
}

func (self *_runtime) evaluateSwitchStatement(node *parser.SwitchStatement) Value {

	labels := append(self.labels, "")
	self.labels = nil

	discriminantResult := self.evaluate(node.Discriminant)
	target := node.Default

	for index, clause := range node.Body {
		test := clause.Test
		if test != nil {
			if self.calculateComparison(token.STRICT_EQUAL, discriminantResult, self.evaluate(test)) {
				target = index
				break
			}
		}
	}

	result := Value{}
	if target != -1 {
		for _, clause := range node.Body[target:] {
			for _, statement := range clause.Consequent {
				value := self.evaluate(statement)
				switch value._valueType {
				case valueResult:
					switch value.evaluateBreak(labels) {
					case resultReturn:
						return value
					case resultBreak:
						return Value{}
					}
				case valueEmpty:
				default:
					result = value
				}
			}
		}
	}

	return result
}
