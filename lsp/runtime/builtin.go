package runtime

import (
	_ "embed"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

//go:embed builtin/target.py
var target []byte

//go:embed builtin/target.py
var helpers []byte


func parseBuiltins(parser *tree_sitter.Parser) {
	targetTree := parser.Parse(target, nil)
	helpersTree := parser.Parse(helpers, nil)

	// TODO: bsena; use a custom cst to validate our existing traversal method
	Traverse(targetTree, func(node *tree_sitter.Node) bool {
		node.Kind()
		// Test if is a type definition
		// identifier
		// and then query?
		node.GrammarName()
		node.IsNamed()
		// start, end := node.ByteRange()
		// v := target[start:end]
		return true
	})

	Traverse(helpersTree, func(node *tree_sitter.Node) bool {
		return true
	})
}
