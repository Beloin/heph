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

type QuerySuite struct {
	suite.Suite
}

func (suite *QuerySuite) newParser() *tree_sitter.Parser {
	parser := tree_sitter.NewParser()

	err := parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_python.Language()))
	suite.Require().NoError(err)

	return parser
}

func (suite *QuerySuite) TestFunctionQuery() {
	parser := suite.newParser()
	pythonTree := parser.Parse(pythonTest, nil)

	symbols, err := query.ExtractFunctions(pythonTree, pythonTest)
	suite.Require().NoError(err)

	suite.Require().NotNil(symbols)
	suite.Require().NotEmpty(symbols)
}

func TestQuerySuite(t *testing.T) {
	suite.Run(t, &QuerySuite{})
}
