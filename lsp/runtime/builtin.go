package runtime

import (
	_ "embed"
	"slices"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

//go:embed builtin/target.py
var target []byte

//go:embed builtin/target.py
var helpers []byte

//go:embed builtin/pybt.py
var pybt []byte

// ParseBuiltins Parses builtin files and extracts their symbols.
// TODO: bsena; Should we use query instead of parsing all symbols?
func ParseBuiltins(parser *tree_sitter.Parser) []*Symbol {
	targetTree := parser.Parse(target, nil)
	// defer targetTree.Close()

	helpersTree := parser.Parse(helpers, nil)
	// defer helpersTree.Close()

	pybtTree := parser.Parse(pybt, nil)
	// defer pybtTree.Close()

	targetSymbols := extractSymbols(targetTree, target)

	helpersSymbols := extractSymbols(helpersTree, helpers)

	pybtSymbols := extractSymbols(pybtTree, pybt)

	return slices.Concat(targetSymbols, helpersSymbols, pybtSymbols)
}

// extractDocument extracts symbols from a given tree and raw byte slice.
func extractSymbols(tree *tree_sitter.Tree, raw []byte) []*Symbol {
	machine := NewMachine(raw)
	state := machine.Start()

	symbols := make([]*Symbol, 0)
	Traverse(tree, func(node *tree_sitter.Node) bool {
		state = state(node)

		if machine.HasSymbol {
			symbolCp := machine.Symbol
			symbols = append(symbols, &symbolCp)
		}

		return true
	})

	return symbols
}
