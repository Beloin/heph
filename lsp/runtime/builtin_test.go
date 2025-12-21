package runtime_test

import (
	"testing"

	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/stretchr/testify/suite"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

var allDefs = []string{
	// From helpers.py
	"text_file", "json_file", "tool_target", "group",

	// From pybt.py - Top-level functions
	"abs", "any", "all", "bool", "chr", "dict", "dir",
	"enumerate", "fail", "float", "getattr", "hasattr",
	"hash", "int", "len", "list", "max", "min", "ord",
	"print", "range", "repr", "reversed", "set", "sorted",
	"str", "tuple", "type", "zip",

	// TODO: bsena; extract class

	// From pybt.py - Dict class methods
	"clear", "get", "items", "keys", "pop", "popitem",
	"setdefault", "update", "values",

	// From pybt.py - List class methods
	"append", "clear", "extend", "index", "insert", "pop", "remove",

	// From pybt.py - Set class methods
	"union",

	// From pybt.py - String class methods
	"elem_ords", "capitalize", "codepoint_ords", "count",
	"endswith", "find", "format", "index", "isalnum",
	"isalpha", "isdigit", "islower", "isspace", "istitle",
	"isupper", "join", "lower", "lstrip", "partition",
	"removeprefix", "removesuffix", "replace", "rfind",
	"rindex", "rpartition", "rsplit", "rstrip", "split",
	"elems", "codepoints", "splitlines", "startswith",
	"strip", "title", "upper",

	// From target.py
	"target",
}

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

	// Extract symbol names for comparison
	symbolNames := make([]string, 0, len(symbols))
	for _, sym := range symbols {
		symbolNames = append(symbolNames, sym.Name)
	}

	suite.Require().ElementsMatch(allDefs, symbolNames)
}

func TestTBuiltinSuite(t *testing.T) {
	suite.Run(t, &BuiltinSuite{})
}
