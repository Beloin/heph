package runtime_test

import (
	_ "embed"
	"testing"

	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/stretchr/testify/suite"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

//go:embed testdata/test.py
var pythonTest []byte

type TraverseSuite struct {
	suite.Suite
}

func (suite *TraverseSuite) newParser() *tree_sitter.Parser {
	parser := tree_sitter.NewParser()

	err := parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_python.Language()))
	suite.Require().NoError(err)

	return parser
}

func (suite *TraverseSuite) TestPySimpleFile() {
	parser := suite.newParser()
	targetTree := parser.Parse(pythonTest, nil)

	suite.Require().NotNil(targetTree)
}

func (suite *TraverseSuite) TestPyMachine() {
	parser := suite.newParser()
	targetTree := parser.Parse(pythonTest, nil)

	functionName := ""
	machine := runtime.NewMachine()
	state := machine.Start()
	runtime.Traverse(targetTree, func(node *tree_sitter.Node) bool {
		state = state(node.Kind())

		if machine.HasSymbol {
			start, end := node.ByteRange()
			v := pythonTest[start:end]
			functionName = string(v)

			return false
		}

		return true
	})

	suite.Require().Equal("my_custom_function", functionName)
}

func TestTraverseSuite(t *testing.T) {
	suite.Run(t, &TraverseSuite{})
}
