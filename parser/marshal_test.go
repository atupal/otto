package parser

import (
	"encoding/json"
	"reflect"
)

func marshal(name string, children ...interface{}) map[string]interface{} {
	if len(children) == 1 {
		return map[string]interface{}{
			name: children[0],
		}
	}
	map_ := map[string]interface{}{}
	length := len(children) / 2
	for i := 0; i < length; i++ {
		name := children[i*2].(string)
		value := children[i*2+1]
		map_[name] = value
	}
	if name == "" {
		return map_
	}
	return map[string]interface{}{
		name: map_,
	}
}

func testMarshalNode(node interface{}) interface{} {
	switch node := node.(type) {

	// Expression

	case *ArrayLiteral:
		return marshal("Array", testMarshalNode(node.Value))

	case *AssignExpression:
		return marshal("Assign",
			"Left", testMarshalNode(node.Left),
			"Right", testMarshalNode(node.Right),
		)

	case *BinaryExpression:
		return marshal("BinaryExpression",
			"Operator", node.Operator.String(),
			"Left", testMarshalNode(node.Left),
			"Right", testMarshalNode(node.Right),
		)

	case *BooleanLiteral:
		return marshal("Literal", node.Value)

	case *CallExpression:
		return marshal("Call",
			"Callee", testMarshalNode(node.Callee),
			"ArgumentList", testMarshalNode(node.ArgumentList),
		)

	case *ConditionalExpression:
		return marshal("Conditional",
			"Test", testMarshalNode(node.Test),
			"Consequent", testMarshalNode(node.Consequent),
			"Alternate", testMarshalNode(node.Alternate),
		)

	case *DotExpression:
		return marshal("Dot",
			"Left", testMarshalNode(node.Left),
			"Member", node.Identifier.Name,
		)

	case *NewExpression:
		return marshal("New",
			"Callee", testMarshalNode(node.Callee),
			"ArgumentList", testMarshalNode(node.ArgumentList),
		)

	case *NullLiteral:
		return marshal("Literal", nil)

	case *NumberLiteral:
		return marshal("Literal", node.Value)

	case *ObjectLiteral:
		return marshal("Object", testMarshalNode(node.Value))

	case *RegExpLiteral:
		return marshal("Literal", node.Literal)

	case *StringLiteral:
		return marshal("Literal", node.Literal)

	// Statement

	case *Program:
		return testMarshalNode(node.Body)

	case *BlockStatement:
		return marshal("BlockStatement", testMarshalNode(node.List))

	case *EmptyStatement:
		return "EmptyStatement"

	case *ExpressionStatement:
		return testMarshalNode(node.Expression)

	case *FunctionExpression:
		return marshal("Function", testMarshalNode(node.Body))

	case *Identifier:
		return marshal("Identifier", node.Name)

	case *IfStatement:
		if_ := marshal("",
			"Test", testMarshalNode(node.Test),
			"Consequent", testMarshalNode(node.Consequent),
		)
		if node.Alternate != nil {
			if_["Alternate"] = testMarshalNode(node.Alternate)
		}
		return marshal("If", if_)

	case *LabelledStatement:
		return marshal("Label",
			"Name", node.Label.Name,
			"Statement", testMarshalNode(node.Statement),
		)
	case Property:
		return marshal("",
			"Key", node.Key,
			"Value", testMarshalNode(node.Value),
		)

	case *ReturnStatement:
		return marshal("Return", testMarshalNode(node.Argument))

	case *SequenceExpression:
		return marshal("Sequence", testMarshalNode(node.Sequence))
	}

	{
		value := reflect.ValueOf(node)
		if value.Kind() == reflect.Slice {
			tmp0 := []interface{}{}
			for index := 0; index < value.Len(); index++ {
				tmp0 = append(tmp0, testMarshalNode(value.Index(index).Interface()))
			}
			return tmp0
		}
	}

	return nil
}

func testMarshal(node interface{}) string {
	value, err := json.Marshal(testMarshalNode(node))
	if err != nil {
		panic(err)
	}
	return string(value)
}
