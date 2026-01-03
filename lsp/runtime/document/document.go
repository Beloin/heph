package document

import (
	"sync"

	"github.com/hephbuild/heph/lsp/runtime/query"
	"github.com/hephbuild/heph/lsp/runtime/symbol"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type Document struct {
	Name string

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

func NewDocument(name string, tree *tree_sitter.Tree, rawText []byte) (*Document, error) {
	doc := &Document{Name: name, Tree: tree, Text: rawText, TextString: string(rawText)}

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
	symbols, err := query.QuerySymbols(d.Tree, d.Text, d.Name)
	if err != nil {
		return err
	}

	// TODO: bsena; also extract targets here?
	d.Symbols = symbols

	return nil
}

func (d *Document) ExtractCurrentStringLiteral(byteOffSet uint) string {
	return query.ExtractCurrentStringLiteral(d.Tree.RootNode(), d.Text, byteOffSet)
}

func (d *Document) ExtractCurrentSymbol(byteOffSet uint) (*symbol.Symbol, bool) {
	sName := d.ExtractCurrentSymbolName(byteOffSet)
	return d.Query(sName)
}

func (d *Document) ExtractCurrentSymbolName(byteOffSet uint) string {
	return query.ExtractCurrentSymbol(d.Tree.RootNode(), d.Text, byteOffSet)
}

func (d *Document) ExtractCurrentFunctionName(byteOffSet uint) string {
	return query.ExtractFunctionNameFromOffset(d.Tree.RootNode(), d.Text, byteOffSet)
}

func (d *Document) Query(symbolName string) (*symbol.Symbol, bool) {
	return symbol.FindSymbol(d.Symbols, symbolName)
}
