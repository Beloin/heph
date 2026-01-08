package query_test

import (
	_ "embed"
	"testing"

	"github.com/hephbuild/heph/lsp/runtime/query"
	"github.com/stretchr/testify/suite"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

//go:embed testdata/call_test.py
var callTest []byte

var (
	expectedFunctionCalls = []string{"load", "print", "my_other_call"}
)

type CallQuerySuite struct {
	suite.Suite
}

func (suite *CallQuerySuite) newParser() *tree_sitter.Parser {
	parser := tree_sitter.NewParser()

	err := parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_python.Language()))
	suite.Require().NoError(err)

	return parser
}

func (suite *CallQuerySuite) TestFunctionCallQuery() {
	parser := suite.newParser()
	pythonTree := parser.Parse(callTest, nil)

	symbols, err := query.QueryCalls(pythonTree, callTest, "")
	suite.Require().NoError(err)

	// Extract function call names
	callNames := []string{}
	for _, s := range symbols {
		callNames = append(callNames, s.Name)
	}

	suite.Require().NotNil(symbols)
	suite.Require().NotEmpty(symbols)
	suite.Require().ElementsMatch(expectedFunctionCalls, callNames)
}

func TestCallQuerySuite(t *testing.T) {
	suite.Run(t, &CallQuerySuite{})
}
