package runtime_test

import (
	"testing"

	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/stretchr/testify/suite"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

type BuiltinSuite struct {
	suite.Suite
}

func (suite *BuiltinSuite) newParser() *tree_sitter.Parser {
	parser := tree_sitter.NewParser()

	err := parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_python.Language()))
	suite.Require().NoError(err)

	return parser
}

func (suite *BuiltinSuite) TestSymbols() {
	parser := suite.newParser()
	symbols := runtime.ParseBuiltins(parser)

	suite.Require().NotNil(symbols)
	suite.Require().NotEmpty(symbols)
}

func TestTBuiltinSuite(t *testing.T) {
	suite.Run(t, &BuiltinSuite{})
}
