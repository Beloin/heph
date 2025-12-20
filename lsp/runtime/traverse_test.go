package runtime_test

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/stretchr/testify/suite"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

//go:embed testdata/test.py
var pythonTest []byte

var pythonNames = []string{"my_custom_function", "my_custom_variable", "my_custom_result"}

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

// TODO: bsena; After all our tests, remove this logs
func (suite *TraverseSuite) TestPyMachine() {
	parser := suite.newParser()
	targetTree := parser.Parse(pythonTest, nil)

	functionName := ""
	machine := runtime.NewMachine(pythonTest)
	state := machine.Start()
	runtime.Traverse(targetTree, func(node *tree_sitter.Node) bool {
		state = state(node)
		// fmt.Println(node.Kind())

		if machine.HasSymbol {
			fmt.Printf("Custom Type: %s, Name: %s, Value: %s\n", machine.TSSymbolKind, machine.SymbolName, machine.SymbolValue)
		}

		if machine.HasSymbol && machine.IsFunction {
			functionName = machine.SymbolName

			// return false
		}

		return true
	})

	suite.Require().Equal("my_custom_function", functionName)
}

func (suite *TraverseSuite) TestPyNames() {
	parser := suite.newParser()
	targetTree := parser.Parse(pythonTest, nil)

	machine := runtime.NewMachine(pythonTest)
	state := machine.Start()

	names := []string{}

	runtime.Traverse(targetTree, func(node *tree_sitter.Node) bool {
		state = state(node)
		if machine.HasSymbol {
			names = append(names, machine.SymbolName)
		}

		return true
	})

	suite.Require().ElementsMatch(pythonNames, names)
}

func TestTraverseSuite(t *testing.T) {
	suite.Run(t, &TraverseSuite{})
}
