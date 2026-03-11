package builtin

import (
	_ "embed"
	"slices"

	"github.com/hephbuild/heph/internal/hlsp/runtime/query"
	"github.com/hephbuild/heph/internal/hlsp/runtime/symbol"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

// TODO: bsena; This will all be changed, we should use the plugin layer using the "server" connection

//go:embed target2.py
var target []byte

//go:embed helpers.py
var helpers []byte

//go:embed pybt.py
var pybt []byte

const builtinSource = "hbuiltins"

// TODO: bsena; We will actually create builtins by "hand"
// The pybt can be still loaded from the python files
// but starlark should be manually created

var (
	targetSymbols  []*symbol.Symbol
	helpersSymbols []*symbol.Symbol
	pybtSymbols    []*symbol.Symbol
)

// InitBuiltins parses all builtin stub files and stores the results in package-level vars.
// Must be called once before using any getter.
func InitBuiltins(parser *tree_sitter.Parser) error {
	targetTree := parser.Parse(target, nil)
	defer targetTree.Close()

	helpersTree := parser.Parse(helpers, nil)
	defer helpersTree.Close()

	pybtTree := parser.Parse(pybt, nil)
	defer pybtTree.Close()

	var err error

	targetSymbols, err = query.QuerySymbols(targetTree, target, builtinSource, nil)
	if err != nil {
		return err
	}

	helpersSymbols, err = query.QuerySymbols(helpersTree, helpers, builtinSource, nil)
	if err != nil {
		return err
	}

	pybtSymbols, err = query.QuerySymbols(pybtTree, pybt, builtinSource, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetTarget returns the symbols parsed from target2.py (the target() builtin).
func GetTarget() []*symbol.Symbol {
	return targetSymbols
}

// GetHelpers returns the symbols parsed from helpers.py (load, group, text_file, …).
func GetHelpers() []*symbol.Symbol {
	return helpersSymbols
}

// GetSKBuiltins returns the symbols parsed from pybt.py (Starlark built-ins: abs, any, len, …).
func GetSKBuiltins() []*symbol.Symbol {
	return pybtSymbols
}

// All returns all builtin symbols concatenated.
func All() []*symbol.Symbol {
	return slices.Concat(targetSymbols, helpersSymbols, pybtSymbols)
}
