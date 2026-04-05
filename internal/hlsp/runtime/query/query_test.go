package query_test

import (
	_ "embed"
	"testing"

	"github.com/hephbuild/heph/internal/hlsp/runtime/query"
	"github.com/hephbuild/heph/internal/hlsp/runtime/symbol"
	"github.com/stretchr/testify/suite"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

//go:embed testdata/test.py
var pythonTest2 []byte

var (
	functionNames2 = []string{"my_custom_function", "my_other_function", "my_argless_function", "my_documented_function", "level1"}
	testVariables2 = []string{"my_custom_variable", "my_new_var", "my_custom_result"}
)

type Query2Suite struct {
	suite.Suite
}

func (suite *Query2Suite) newParser() *tree_sitter.Parser {
	parser := tree_sitter.NewParser()

	err := parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_python.Language()))
	suite.Require().NoError(err)

	return parser
}

func (suite *Query2Suite) TestQueryAll() {
	parser := suite.newParser()
	pythonTree := parser.Parse(pythonTest2, nil)

	root, err := query.QueryAll(pythonTree, pythonTest2, "")
	suite.Require().NoError(err)

	suite.Require().NotNil(root)
	suite.Require().Equal(symbol.RootKind, root.Kind)
}

func (suite *Query2Suite) TestFunctions() {
	parser := suite.newParser()
	pythonTree := parser.Parse(pythonTest2, nil)

	root, err := query.QueryAll(pythonTree, pythonTest2, "")
	suite.Require().NoError(err)

	functions := filterSymbolsByKind(root.Symbols, symbol.FunctionKind)
	names := []string{}
	for _, s := range functions {
		names = append(names, s.Name)
	}

	suite.Require().NotNil(functions)
	suite.Require().NotEmpty(functions)
	suite.Require().ElementsMatch(functionNames2, names)
}

func (suite *Query2Suite) TestFunctionParameters() {
	parser := suite.newParser()
	pythonTree := parser.Parse(pythonTest2, nil)

	root, err := query.QueryAll(pythonTree, pythonTest2, "")
	suite.Require().NoError(err)

	functions := filterSymbolsByKind(root.Symbols, symbol.FunctionKind)

	var customFunc, otherFunc *symbol.Symbol
	for _, s := range functions {
		switch s.Name {
		case "my_custom_function":
			customFunc = s
		case "my_other_function":
			otherFunc = s
		}
	}

	suite.Require().NotNil(customFunc, "my_custom_function should be found")
	suite.Require().NotNil(otherFunc, "my_other_function should be found")

	suite.Require().Len(customFunc.Parameters, 2, "my_custom_function should have 2 parameters")
	customFuncParamNames := []string{}
	for _, param := range customFunc.Parameters {
		customFuncParamNames = append(customFuncParamNames, param.Name)
	}
	suite.Require().Contains(customFuncParamNames, "arg1", "my_custom_function should have arg1 parameter")
	suite.Require().Contains(customFuncParamNames, "arg2", "my_custom_function should have arg2 parameter")

	suite.Require().Len(otherFunc.Parameters, 4, "my_other_function should have 4 parameters")
	var otherFuncParamNames []string
	for _, param := range otherFunc.Parameters {
		otherFuncParamNames = append(otherFuncParamNames, param.Name)
	}
	suite.Require().Contains(otherFuncParamNames, "arg1", "my_other_function should have arg1 parameter")
	suite.Require().Contains(otherFuncParamNames, "arg2", "my_other_function should have arg2 parameter")
	suite.Require().Contains(otherFuncParamNames, "arg3", "my_other_function should have arg3 parameter")
	suite.Require().Contains(otherFuncParamNames, "arg4", "my_other_function should have arg4 parameter")
}

func (suite *Query2Suite) TestVariables() {
	parser := suite.newParser()
	pythonTree := parser.Parse(pythonTest2, nil)

	root, err := query.QueryAll(pythonTree, pythonTest2, "")
	suite.Require().NoError(err)

	variables := filterSymbolsByKind(root.Symbols, symbol.VariableKind)
	names := []string{}
	for _, s := range variables {
		names = append(names, s.Name)
	}

	suite.Require().NotNil(variables)
	suite.Require().NotEmpty(variables)
	suite.Require().ElementsMatch(testVariables2, names)
}

func (suite *Query2Suite) TestFunctionArgsDoc() {
	parser := suite.newParser()
	pythonTree := parser.Parse(pythonTest2, nil)

	root, err := query.QueryAll(pythonTree, pythonTest2, "")
	suite.Require().NoError(err)

	functions := filterSymbolsByKind(root.Symbols, symbol.FunctionKind)

	var documentedFunc *symbol.Symbol
	for _, s := range functions {
		if s.Name == "my_documented_function" {
			documentedFunc = s
			break
		}
	}
	suite.Require().NotNil(documentedFunc, "my_documented_function should be found")

	suite.Require().Len(documentedFunc.Parameters, 2, "my_documented_function should have 2 parameters")

	suite.Require().Equal("param1", documentedFunc.Parameters[0].Name)
	suite.Require().Equal("str", documentedFunc.Parameters[0].Type.Name)
	suite.Require().Equal("The first parameter.", documentedFunc.Parameters[0].DocString)

	suite.Require().Equal("param2", documentedFunc.Parameters[1].Name)
	suite.Require().Equal("int", documentedFunc.Parameters[1].Type.Name)
	suite.Require().Equal("The second parameter. Defaults to 0.", documentedFunc.Parameters[1].DocString)
	suite.Require().NotNil(documentedFunc.Parameters[1].Value, "Parameter default value should be resolved")
	suite.Require().Equal("0", documentedFunc.Parameters[1].Value.Value, "Default value should be '0'")
}

func (suite *Query2Suite) TestNestedFunctions() {
	parser := suite.newParser()
	pythonTree := parser.Parse(pythonTest2, nil)

	root, err := query.QueryAll(pythonTree, pythonTest2, "")
	suite.Require().NoError(err)

	functions := filterSymbolsByKind(root.Symbols, symbol.FunctionKind)

	var level1 *symbol.Symbol
	for _, s := range functions {
		if s.Name == "level1" {
			level1 = s
			break
		}
	}
	suite.Require().NotNil(level1, "level1 should be a top-level function")

	level2 := findSymbolInSlice(level1.Symbols, "level2")
	suite.Require().NotNil(level2, "level2 should be nested in level1")

	level3 := findSymbolInSlice(level2.Symbols, "level3")
	suite.Require().NotNil(level3, "level3 should be nested in level2")

	level4 := findSymbolInSlice(level3.Symbols, "level4")
	suite.Require().NotNil(level4, "level4 should be nested in level3")
}

func (suite *Query2Suite) TestCalls() {
	parser := suite.newParser()
	pythonTree := parser.Parse(pythonTest2, nil)

	root, err := query.QueryAll(pythonTree, pythonTest2, "")
	suite.Require().NoError(err)

	calls := filterSymbolsByKind(root.Symbols, symbol.FunctionCallKind)

	suite.Require().NotNil(calls)
	suite.Require().NotEmpty(calls)

	callNames := []string{}
	for _, c := range calls {
		callNames = append(callNames, c.Name)
	}

	suite.Require().Contains(callNames, "my_custom_function")
	suite.Require().Contains(callNames, "print")
}

func (suite *Query2Suite) TestTypeInferenceFromParent() {
	parser := suite.newParser()

	code := []byte(`x = 1
y = x
`)
	pythonTree := parser.Parse(code, nil)

	root, err := query.QueryAll(pythonTree, code, "test")
	suite.Require().NoError(err)

	variables := filterSymbolsByKind(root.Symbols, symbol.VariableKind)
	suite.Require().Len(variables, 2)

	// Find x and y
	var xSym, ySym *symbol.Symbol
	for _, v := range variables {
		if v.Name == "x" {
			xSym = v
		}
		if v.Name == "y" {
			ySym = v
		}
	}

	suite.Require().NotNil(xSym)
	suite.Require().NotNil(ySym)

	// x should have type int
	suite.Require().NotNil(xSym.Type)
	suite.Require().Equal(symbol.PrimitiveInt, xSym.Type)

	// y should reference x
	suite.Require().NotNil(ySym.Type)
	suite.Require().Equal(xSym, ySym.Type, "y's type should point to x's symbol")
}

func (suite *Query2Suite) TestParentChain_RootSymbols() {
	parser := suite.newParser()

	code := []byte(`x = 1
y = 2

def foo():
    pass
`)
	pythonTree := parser.Parse(code, nil)

	root, err := query.QueryAll(pythonTree, code, "test")
	suite.Require().NoError(err)

	// Root symbols should have root as parent
	variables := filterSymbolsByKind(root.Symbols, symbol.VariableKind)
	suite.Require().Len(variables, 2)

	for _, v := range variables {
		suite.Require().NotNil(v.Parent, "Variable %s should have parent", v.Name)
		suite.Require().Equal(root, v.Parent, "Variable %s's parent should be root", v.Name)
	}

	functions := filterSymbolsByKind(root.Symbols, symbol.FunctionKind)
	suite.Require().Len(functions, 1)

	foo := functions[0]
	suite.Require().NotNil(foo.Parent, "foo should have parent")
	suite.Require().Equal(root, foo.Parent, "foo's parent should be root")
}

func (suite *Query2Suite) TestParentChain_NestedFunctions() {
	parser := suite.newParser()

	code := []byte(`def outer():
    def inner():
        pass
    pass
`)
	pythonTree := parser.Parse(code, nil)

	root, err := query.QueryAll(pythonTree, code, "test")
	suite.Require().NoError(err)

	functions := filterSymbolsByKind(root.Symbols, symbol.FunctionKind)
	suite.Require().Len(functions, 1)

	outer := functions[0]
	suite.Require().NotNil(outer.Parent, "outer should have parent")
	suite.Require().Equal(root, outer.Parent, "outer's parent should be root")

	// inner should be in outer's Symbols
	suite.Require().Len(outer.Symbols, 1, "outer should have 1 nested symbol")

	inner := outer.Symbols[0]
	suite.Require().Equal("inner", inner.Name)
	suite.Require().NotNil(inner.Parent, "inner should have parent")
	suite.Require().Equal(outer, inner.Parent, "inner's parent should be outer")
}

func (suite *Query2Suite) TestParentChain_VariablesInFunction() {
	parser := suite.newParser()

	code := []byte(`def foo():
    x = 1
    y = x
`)
	pythonTree := parser.Parse(code, nil)

	root, err := query.QueryAll(pythonTree, code, "test")
	suite.Require().NoError(err)

	functions := filterSymbolsByKind(root.Symbols, symbol.FunctionKind)
	suite.Require().Len(functions, 1)

	foo := functions[0]
	suite.Require().NotNil(foo.Parent, "foo should have parent")
	suite.Require().Equal(root, foo.Parent, "foo's parent should be root")

	// Variables inside foo
	variables := filterSymbolsByKind(foo.Symbols, symbol.VariableKind)
	suite.Require().Len(variables, 2)

	for _, v := range variables {
		suite.Require().NotNil(v.Parent, "Variable %s should have parent", v.Name)
		suite.Require().Equal(foo, v.Parent, "Variable %s's parent should be foo", v.Name)
	}

	// Find x and y
	var xSym, ySym *symbol.Symbol
	for _, v := range variables {
		if v.Name == "x" {
			xSym = v
		}
		if v.Name == "y" {
			ySym = v
		}
	}

	suite.Require().NotNil(xSym)
	suite.Require().NotNil(ySym)

	// y should reference x (type inference across same scope)
	suite.Require().NotNil(ySym.Type, "y should have a type")
	suite.Require().Equal(xSym, ySym.Type, "y's type should reference x")
}

func (suite *Query2Suite) TestParentChain_DeeplyNestedFunctions() {
	parser := suite.newParser()

	code := []byte(`def level1():
    def level2():
        def level3():
            x = 1
        pass
    pass
`)
	pythonTree := parser.Parse(code, nil)

	root, err := query.QueryAll(pythonTree, code, "test")
	suite.Require().NoError(err)

	// level1
	level1Funcs := filterSymbolsByKind(root.Symbols, symbol.FunctionKind)
	suite.Require().Len(level1Funcs, 1)

	level1 := level1Funcs[0]
	suite.Require().NotNil(level1.Parent)
	suite.Require().Equal(root, level1.Parent, "level1's parent should be root")

	// level2
	suite.Require().Len(level1.Symbols, 1)
	level2 := level1.Symbols[0]
	suite.Require().Equal("level2", level2.Name)
	suite.Require().NotNil(level2.Parent)
	suite.Require().Equal(level1, level2.Parent, "level2's parent should be level1")

	// level3
	suite.Require().Len(level2.Symbols, 1)
	level3 := level2.Symbols[0]
	suite.Require().Equal("level3", level3.Name)
	suite.Require().NotNil(level3.Parent)
	suite.Require().Equal(level2, level3.Parent, "level3's parent should be level2")

	// x inside level3
	suite.Require().Len(level3.Symbols, 1)
	x := level3.Symbols[0]
	suite.Require().Equal("x", x.Name)
	suite.Require().NotNil(x.Parent)
	suite.Require().Equal(level3, x.Parent, "x's parent should be level3")
}

func (suite *Query2Suite) TestParentChain_MixedNesting() {
	parser := suite.newParser()

	code := []byte(`root_var = 1

def outer():
    outer_var = root_var
    
    def inner():
        inner_var = outer_var
    pass
`)
	pythonTree := parser.Parse(code, nil)

	root, err := query.QueryAll(pythonTree, code, "test")
	suite.Require().NoError(err)

	// Root level
	rootVars := filterSymbolsByKind(root.Symbols, symbol.VariableKind)
	suite.Require().Len(rootVars, 1)
	rootVar := rootVars[0]
	suite.Require().Equal(root, rootVar.Parent, "root_var's parent should be root")

	// Outer function
	outerFuncs := filterSymbolsByKind(root.Symbols, symbol.FunctionKind)
	suite.Require().Len(outerFuncs, 1)
	outer := outerFuncs[0]
	suite.Require().Equal(root, outer.Parent, "outer's parent should be root")

	// Outer's variables
	outerVars := filterSymbolsByKind(outer.Symbols, symbol.VariableKind)
	suite.Require().Len(outerVars, 1)
	outerVar := outerVars[0]
	suite.Require().Equal("outer_var", outerVar.Name)
	suite.Require().Equal(outer, outerVar.Parent, "outer_var's parent should be outer")

	// Verify outer_var type references root_var (cross-scope lookup)
	suite.Require().NotNil(outerVar.Type)
	suite.Require().Equal(rootVar, outerVar.Type, "outer_var should reference root_var via type")

	// Inner function
	innerFuncs := filterSymbolsByKind(outer.Symbols, symbol.FunctionKind)
	suite.Require().Len(innerFuncs, 1)
	inner := innerFuncs[0]
	suite.Require().Equal(outer, inner.Parent, "inner's parent should be outer")

	// Inner's variables
	innerVars := filterSymbolsByKind(inner.Symbols, symbol.VariableKind)
	suite.Require().Len(innerVars, 1)
	innerVar := innerVars[0]
	suite.Require().Equal("inner_var", innerVar.Name)
	suite.Require().Equal(inner, innerVar.Parent, "inner_var's parent should be inner")

	// Verify inner_var type references outer_var (cross-scope lookup up the chain)
	suite.Require().NotNil(innerVar.Type)
	suite.Require().Equal(outerVar, innerVar.Type, "inner_var should reference outer_var via type")
}

func filterSymbolsByKind(symbols []*symbol.Symbol, kind symbol.SymbolKind) []*symbol.Symbol {
	var result []*symbol.Symbol
	for _, s := range symbols {
		if s.Kind == kind {
			result = append(result, s)
		}
	}
	return result
}

func TestQuery2Suite(t *testing.T) {
	suite.Run(t, &Query2Suite{})
}

func findSymbolInSlice(syms []*symbol.Symbol, name string) *symbol.Symbol {
	for _, s := range syms {
		if s.Name == name {
			return s
		}
	}
	return nil
}
