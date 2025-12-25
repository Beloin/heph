package document

import (
	"github.com/hephbuild/heph/lsp/runtime/query"
	"github.com/hephbuild/heph/lsp/runtime/symbol"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type Document struct {
	Symbols []*symbol.Symbol
	Tree    *tree_sitter.Tree
	Text    []byte
}

func (d *Document) Close() {
	d.Tree.Close()
}

func NewDocument(tree *tree_sitter.Tree, rawText []byte) (*Document, error) {
	doc := &Document{Tree: tree, Text: rawText}

	symbols, err := query.QuerySymbols(tree, rawText)
	if err != nil {
		return nil, err
	}

	doc.Symbols = symbols

	return doc, nil
}

// SwapTree atomic swaps current tree and return the closed old tree
func (d *Document) SwapTree(newT *tree_sitter.Tree) *tree_sitter.Tree {
	// TODO: bsena; make it atomic with sync.Mutex and update symbols
	old := d.Tree
	old.Close()
	d.Tree = newT

	return old
}
