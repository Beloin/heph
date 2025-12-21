package builtin

import (
	_ "embed"
	"slices"

	"github.com/hephbuild/heph/lsp/runtime/machine"
	"github.com/hephbuild/heph/lsp/runtime/symbol"
	"github.com/hephbuild/heph/lsp/runtime/traverse"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

//go:embed target.py
var target []byte

//go:embed helpers.py
var helpers []byte

//go:embed pybt.py
var pybt []byte

// ParseBuiltins Parses builtin files and extracts their symbols.
// TODO: bsena; Should we use query instead of parsing all symbols?
func ParseBuiltins(parser *tree_sitter.Parser) []*symbol.Symbol {
	targetTree := parser.Parse(target, nil)
	defer targetTree.Close()

	helpersTree := parser.Parse(helpers, nil)
	defer helpersTree.Close()

	pybtTree := parser.Parse(pybt, nil)
	defer pybtTree.Close()

	targetSymbols := ExtractSymbols(targetTree, target)
	helpersSymbols := ExtractSymbols(helpersTree, helpers)
	pybtSymbols := ExtractSymbols(pybtTree, pybt)

	return slices.Concat(targetSymbols, helpersSymbols, pybtSymbols)
}

// ExtractSymbols extracts symbols from a given tree and raw byte slice.
func ExtractSymbols(tree *tree_sitter.Tree, raw []byte) []*symbol.Symbol {
	machine := machine.NewMachine(raw)
	state := machine.Start()

	symbols := make([]*symbol.Symbol, 0)
	traverse.Traverse(tree, func(node *tree_sitter.Node) bool {
		state = state(node)

		if machine.HasSymbol {
			symbolCp := machine.Symbol
			symbols = append(symbols, &symbolCp)
		}

		return true
	})

	return symbols
}
