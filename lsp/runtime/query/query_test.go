package query_test

import (
	_ "embed"
	"testing"

	"github.com/hephbuild/heph/lsp/runtime/query"
	"github.com/stretchr/testify/suite"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

//go:embed testdata/test.py
var pythonTest []byte

//go:embed testdata/class_test.py
var classTest []byte

var (
	functionNames = []string{"my_custom_function", "my_other_function"}
	testVariables = []string{"my_custom_variable", "my_new_var", "my_custom_result"}
	classes       = []string{"MyClass", "MySecondClass"}
	class0Methods = []string{"mymethod", "my_second_method"}
	class1Methods = []string{"method_in_second_class", "static_method_in_class"}
)

type QuerySuite struct {
	suite.Suite
}

func (suite *QuerySuite) newParser() *tree_sitter.Parser {
	parser := tree_sitter.NewParser()

	err := parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_python.Language()))
	suite.Require().NoError(err)

	return parser
}

func (suite *QuerySuite) TestClassQuery() {
	parser := suite.newParser()
	classTree := parser.Parse(classTest, nil)

	symbols, err := query.ExtractClass(classTree, classTest, "")
	suite.Require().NoError(err)
	suite.Require().NotNil(symbols)
	suite.Require().NotEmpty(symbols)

	classesNames := []string{}
	methods := [][]string{}
	for _, s := range symbols {
		classesNames = append(classesNames, s.Name)
		currentMethods := []string{}
		for _, m := range s.Symbols {
			currentMethods = append(currentMethods, m.Name)
		}

		methods = append(methods, currentMethods)
	}

	suite.Require().Len(methods, 2)

	suite.Require().ElementsMatch(classes, classesNames)
	suite.Require().ElementsMatch(class0Methods, methods[0])
	suite.Require().ElementsMatch(class1Methods, methods[1])
}

func (suite *QuerySuite) TestFunctionQuery() {
	parser := suite.newParser()
	pythonTree := parser.Parse(pythonTest, nil)

	symbols, err := query.ExtractFunctions(pythonTree, pythonTest, "")
	suite.Require().NoError(err)

	names := []string{}
	for _, s := range symbols {
		names = append(names, s.Name)
	}

	suite.Require().NotNil(symbols)
	suite.Require().NotEmpty(symbols)
	suite.Require().ElementsMatch(functionNames, names)
}

func (suite *QuerySuite) TestVariablesQuery() {
	parser := suite.newParser()
	pythonTree := parser.Parse(pythonTest, nil)

	symbols, err := query.ExtractVariables(pythonTree, pythonTest, "")
	suite.Require().NoError(err)

	names := []string{}
	for _, s := range symbols {
		names = append(names, s.Name)
	}

	suite.Require().NotNil(symbols)
	suite.Require().NotEmpty(symbols)
	suite.Require().ElementsMatch(testVariables, names)
}

func TestQuerySuite(t *testing.T) {
	suite.Run(t, &QuerySuite{})
}
