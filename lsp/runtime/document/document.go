package document

import (
	"sync"

	"github.com/hephbuild/heph/lsp/runtime/query"
	"github.com/hephbuild/heph/lsp/runtime/symbol"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type Document struct {
	// TODO: bsena; Maybe a tree would be better here
	Symbols    []*symbol.Symbol // TODO: bsena; find a way to index this
	Tree       *tree_sitter.Tree
	Text       []byte // UTF-16 encoded byte array https://microsoft.github.io/language-server-protocol/specifications/specification-3-16/#textDocuments
	TextString string // UTF-16 encoded string https://microsoft.github.io/language-server-protocol/specifications/specification-3-16/#textDocuments

	m sync.Mutex
}

func (d *Document) Close() {
	d.Tree.Close()
}

func NewDocument(tree *tree_sitter.Tree, rawText []byte) (*Document, error) {
	doc := &Document{Tree: tree, Text: rawText, TextString: string(rawText)}

	// TODO: bsena; Extract target names here, look for something like target(name="...")
	err := doc.extractSymbols()

	return doc, err
}

// SwapTree atomic swaps current tree and return the closed old tree
func (d *Document) SwapTree(newT *tree_sitter.Tree, newText []byte) (*tree_sitter.Tree, error) {
	d.m.Lock()
	defer d.m.Unlock()

	oldTree := d.Tree
	oldText := d.Text
	oldTextString := d.TextString

	d.Tree = newT
	d.Text = newText
	d.TextString = string(newText)
	err := d.extractSymbols()
	if err != nil {
		d.Tree = oldTree
		d.Text = oldText
		d.TextString = oldTextString

		return nil, err
	}

	oldTree.Close()

	return oldTree, err
}

func (d *Document) extractSymbols() error {
	symbols, err := query.QuerySymbols(d.Tree, d.Text)
	if err != nil {
		return err
	}

	d.Symbols = symbols

	return nil
}

func (d *Document) Query(symbolName string) (*symbol.Symbol, bool) {
	// TODO: bsena; use a prefix tree and have ALL nodes, even child nodes
	// in that tree
	return findSymbol(d.Symbols, symbolName)
}

func findSymbol(symbols []*symbol.Symbol, sName string) (*symbol.Symbol, bool) {
	for _, symbol := range symbols {
		if symbol.FullName == sName {
			return symbol, true
		}

		if childS, found := findSymbol(symbol.Symbols, sName); found {
			return childS, true
		}
	}

	return nil, false
}
