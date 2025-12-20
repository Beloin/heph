package runtime

import (
	_ "embed"
	"slices"

	protocol "github.com/tliron/glsp/protocol_3_16"
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
func ParseBuiltins(parser *tree_sitter.Parser) []*protocol.DocumentSymbol {
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
func extractSymbols(tree *tree_sitter.Tree, raw []byte) []*protocol.DocumentSymbol {
	machine := NewMachine(raw)
	state := machine.Start()

	symbols := make([]*protocol.DocumentSymbol, 0)
	Traverse(tree, func(node *tree_sitter.Node) bool {
		state = state(node)

		if machine.HasSymbol {
			symbolPos := machine.SymbolPosition
			symbolValue := machine.SymbolValue
			symbol := &protocol.DocumentSymbol{
				Name: machine.SymbolName,
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      protocol.UInteger(symbolPos.RowStart),
						Character: protocol.UInteger(symbolPos.ColumnStart),
					},
					End: protocol.Position{
						Line:      protocol.UInteger(symbolPos.RowEnd),
						Character: protocol.UInteger(symbolPos.ColumnEnd),
					},
				},
				Kind:   MachineKindToProtocolKind(machine.SymbolKind),
				Detail: &symbolValue,
			}
			symbols = append(symbols, symbol)
		}

		return true
	})

	return symbols
}
