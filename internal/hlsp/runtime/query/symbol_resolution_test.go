package query_test

import (
	"testing"

	"github.com/hephbuild/heph/internal/hlsp/runtime/query"
	"github.com/hephbuild/heph/internal/hlsp/runtime/symbol"
	"github.com/stretchr/testify/suite"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

type SymbolResolutionSuite struct {
	suite.Suite
}

func (suite *SymbolResolutionSuite) newParser() *tree_sitter.Parser {
	parser := tree_sitter.NewParser()

	err := parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_python.Language()))
	suite.Require().NoError(err)

	return parser
}

func (suite *SymbolResolutionSuite) TestParameterTypeResolution() {
	parser := suite.newParser()

	code := []byte(`def foo(a: int, b: str, c: list):
    pass
`)
	pythonTree := parser.Parse(code, nil)

	root, err := query.QueryAll(pythonTree, code, "test")
	suite.Require().NoError(err)

	functions := query.FilterSymbolsByKind(root.Symbols, symbol.FunctionKind)
	suite.Require().Len(functions, 1)

	foo := functions[0]
	suite.Require().Len(foo.Parameters, 3)

	// Check int parameter
	suite.Require().Equal("a", foo.Parameters[0].Name)
	suite.Require().NotNil(foo.Parameters[0].Type)
	suite.Require().Equal("int", foo.Parameters[0].Type.Name)
	suite.Require().Equal(symbol.PrimitiveKind, foo.Parameters[0].Type.Kind)

	// Check str parameter
	suite.Require().Equal("b", foo.Parameters[1].Name)
	suite.Require().NotNil(foo.Parameters[1].Type)
	suite.Require().Equal("str", foo.Parameters[1].Type.Name)
	suite.Require().Equal(symbol.PrimitiveKind, foo.Parameters[1].Type.Kind)

	// Check list parameter
	suite.Require().Equal("c", foo.Parameters[2].Name)
	suite.Require().NotNil(foo.Parameters[2].Type)
	suite.Require().Equal("list", foo.Parameters[2].Type.Name)
	suite.Require().Equal(symbol.PrimitiveKind, foo.Parameters[2].Type.Kind)
}

func (suite *SymbolResolutionSuite) TestParameterDefaultValues() {
	parser := suite.newParser()

	code := []byte(`default_val = 42

def foo(a=1, b="hello", c=default_val):
    pass
`)
	pythonTree := parser.Parse(code, nil)

	root, err := query.QueryAll(pythonTree, code, "test")
	suite.Require().NoError(err)

	// Find variables
	variables := query.FilterSymbolsByKind(root.Symbols, symbol.VariableKind)
	suite.Require().Len(variables, 1)
	defaultVar := variables[0]
	suite.Require().Equal("default_val", defaultVar.Name)

	// Find function
	functions := query.FilterSymbolsByKind(root.Symbols, symbol.FunctionKind)
	suite.Require().Len(functions, 1)

	foo := functions[0]
	suite.Require().Len(foo.Parameters, 3)

	// Check integer literal default
	suite.Require().Equal("a", foo.Parameters[0].Name)
	suite.Require().NotNil(foo.Parameters[0].Value)
	suite.Require().Equal("1", foo.Parameters[0].Value.Value)
	suite.Require().Equal(symbol.PrimitiveKind, foo.Parameters[0].Value.Kind)
	suite.Require().Equal(symbol.PrimitiveInt, foo.Parameters[0].Value.Type)

	// Check string literal default
	suite.Require().Equal("b", foo.Parameters[1].Name)
	suite.Require().NotNil(foo.Parameters[1].Value)
	suite.Require().Equal("\"hello\"", foo.Parameters[1].Value.Value)
	suite.Require().Equal(symbol.PrimitiveKind, foo.Parameters[1].Value.Kind)
	suite.Require().Equal(symbol.PrimitiveString, foo.Parameters[1].Value.Type)

	// Check variable reference default - should resolve to actual symbol
	suite.Require().Equal("c", foo.Parameters[2].Name)
	suite.Require().NotNil(foo.Parameters[2].Value)
	suite.Require().Equal("default_val", foo.Parameters[2].Value.Name)
	suite.Require().Equal(symbol.VariableKind, foo.Parameters[2].Value.Kind)
	// Should point to same symbol instance
	suite.Require().Equal(defaultVar, foo.Parameters[2].Value, "Should reference the same symbol from scope")
}

func (suite *SymbolResolutionSuite) TestCallArgumentResolution() {
	parser := suite.newParser()

	code := []byte(`x = 10
y = "hello"

func_call(x)
func_call(y)
func_call(42)
func_call("literal")
`)
	pythonTree := parser.Parse(code, nil)

	root, err := query.QueryAll(pythonTree, code, "test")
	suite.Require().NoError(err)

	// Find variables
	variables := query.FilterSymbolsByKind(root.Symbols, symbol.VariableKind)
	var xVar, yVar *symbol.Symbol
	for _, v := range variables {
		if v.Name == "x" {
			xVar = v
		}
		if v.Name == "y" {
			yVar = v
		}
	}
	suite.Require().NotNil(xVar)
	suite.Require().NotNil(yVar)

	// Find calls
	calls := query.FilterSymbolsByKind(root.Symbols, symbol.FunctionCallKind)
	suite.Require().Len(calls, 4)

	// First call: func_call(x) - should resolve to variable
	call1 := calls[0]
	suite.Require().Len(call1.Parameters, 1)
	suite.Require().NotNil(call1.Parameters[0].Value)
	suite.Require().Equal("x", call1.Parameters[0].Value.Name)
	suite.Require().Equal(symbol.VariableKind, call1.Parameters[0].Value.Kind)
	suite.Require().Equal(xVar, call1.Parameters[0].Value, "Should reference same x variable")

	// Second call: func_call(y) - should resolve to variable
	call2 := calls[1]
	suite.Require().Len(call2.Parameters, 1)
	suite.Require().NotNil(call2.Parameters[0].Value)
	suite.Require().Equal("y", call2.Parameters[0].Value.Name)
	suite.Require().Equal(symbol.VariableKind, call2.Parameters[0].Value.Kind)
	suite.Require().Equal(yVar, call2.Parameters[0].Value, "Should reference same y variable")

	// Third call: func_call(42) - should be int literal
	call3 := calls[2]
	suite.Require().Len(call3.Parameters, 1)
	suite.Require().NotNil(call3.Parameters[0].Value)
	suite.Require().Equal("42", call3.Parameters[0].Value.Value)
	suite.Require().Equal(symbol.PrimitiveKind, call3.Parameters[0].Value.Kind)
	suite.Require().Equal(symbol.PrimitiveInt, call3.Parameters[0].Value.Type)

	// Fourth call: func_call("literal") - should be string literal
	call4 := calls[3]
	suite.Require().Len(call4.Parameters, 1)
	suite.Require().NotNil(call4.Parameters[0].Value)
	suite.Require().Equal("\"literal\"", call4.Parameters[0].Value.Value)
	suite.Require().Equal(symbol.PrimitiveKind, call4.Parameters[0].Value.Kind)
	suite.Require().Equal(symbol.PrimitiveString, call4.Parameters[0].Value.Type)
}

func (suite *SymbolResolutionSuite) TestLiteralSymbolKinds() {
	parser := suite.newParser()

	code := []byte(`def foo(
	a=123,
	b=45.67,
	c=True,
	d=False,
	e="text",
	f=[1,2,3],
	g={"key": "val"}
):
    pass
`)
	pythonTree := parser.Parse(code, nil)

	root, err := query.QueryAll(pythonTree, code, "test")
	suite.Require().NoError(err)

	functions := query.FilterSymbolsByKind(root.Symbols, symbol.FunctionKind)
	suite.Require().Len(functions, 1)

	foo := functions[0]
	suite.Require().Len(foo.Parameters, 7)

	// Test each literal kind
	suite.Require().Equal(symbol.PrimitiveKind, foo.Parameters[0].Value.Kind)
	suite.Require().Equal(symbol.PrimitiveInt, foo.Parameters[0].Value.Type)
	suite.Require().Equal("123", foo.Parameters[0].Value.Value)

	suite.Require().Equal(symbol.PrimitiveKind, foo.Parameters[1].Value.Kind)
	suite.Require().Equal(symbol.PrimitiveFloat, foo.Parameters[1].Value.Type)
	suite.Require().Equal("45.67", foo.Parameters[1].Value.Value)

	suite.Require().Equal(symbol.PrimitiveKind, foo.Parameters[2].Value.Kind)
	suite.Require().Equal(symbol.PrimitiveBool, foo.Parameters[2].Value.Type)
	suite.Require().Equal("True", foo.Parameters[2].Value.Value)

	suite.Require().Equal(symbol.PrimitiveKind, foo.Parameters[3].Value.Kind)
	suite.Require().Equal(symbol.PrimitiveBool, foo.Parameters[3].Value.Type)
	suite.Require().Equal("False", foo.Parameters[3].Value.Value)

	suite.Require().Equal(symbol.PrimitiveKind, foo.Parameters[4].Value.Kind)
	suite.Require().Equal(symbol.PrimitiveString, foo.Parameters[4].Value.Type)
	suite.Require().Equal("\"text\"", foo.Parameters[4].Value.Value)

	suite.Require().Equal(symbol.PrimitiveKind, foo.Parameters[5].Value.Kind)
	suite.Require().Equal(symbol.ListType, foo.Parameters[5].Value.Type)

	suite.Require().Equal(symbol.PrimitiveKind, foo.Parameters[6].Value.Kind)
	suite.Require().Equal(symbol.DictType, foo.Parameters[6].Value.Type)
}

func TestSymbolResolutionSuite(t *testing.T) {
	suite.Run(t, &SymbolResolutionSuite{})
}
