package document

import (
	"strings"
	"sync"

	"github.com/hephbuild/heph/lsp/runtime/builtin"
	"github.com/hephbuild/heph/lsp/runtime/query"
	"github.com/hephbuild/heph/lsp/runtime/symbol"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type Document struct {
	// FullPath
	FullPath string

	// TODO: bsena; Maybe a tree would be better here
	Symbols  []*symbol.Symbol // TODO: bsena; find a way to index this
	Calls    []*symbol.Symbol
	Loads    []string // TODO: bsena; Use the graph so we can know where to load thinks
	DocLoads []*Document
	// ExportedTargets []*specs.Target // TODO: bsena; We need the DAG from heph

	Tree       *tree_sitter.Tree
	Text       []byte // UTF-16 encoded byte array https://microsoft.github.io/language-server-protocol/specifications/specification-3-16/#textDocuments
	TextString string // UTF-16 encoded string https://microsoft.github.io/language-server-protocol/specifications/specification-3-16/#textDocuments

	m sync.Mutex
}

func (d *Document) Close() {
	d.Tree.Close()
}

func NewDocument(name string, tree *tree_sitter.Tree, rawText []byte) (*Document, error) {
	doc := &Document{FullPath: name, Tree: tree, Text: rawText, TextString: string(rawText)}

	// TODO: bsena; Extract target names here, look for something like target(name="...")
	syms, calls, err := extractSymbols(doc.Tree, doc.Text, doc.FullPath)
	doc.Symbols = syms
	doc.Calls = calls
	doc.extractLoads() // TODO: bsena; find a way to do this whitin the same method

	return doc, err
}

// SwapTree atomic swaps current tree and return the closed old tree
func (d *Document) SwapTree(newT *tree_sitter.Tree, newText []byte) (*tree_sitter.Tree, error) {
	d.m.Lock()
	defer d.m.Unlock()

	oldTree := d.Tree

	syms, calls, err := extractSymbols(newT, newText, d.FullPath)
	if err != nil {
		return nil, err
	}

	d.Symbols = syms
	d.Calls = calls
	d.Tree = newT
	d.Text = newText
	d.TextString = string(newText)
	d.extractLoads() // TODO: bsena; find a way to do this whitin the same method
	oldTree.Close()

	return oldTree, err
}

// TODO: bsena; also extract targets here?
// So we can have a custom symbol that is a spec.Target?
// extractSymbols reads all symbols from document
// returns document symbols, document calls and error
func extractSymbols(tree *tree_sitter.Tree, text []byte, source string) ([]*symbol.Symbol, []*symbol.Symbol, error) {
	symbols, err := query.QuerySymbols(tree, text, source)
	if err != nil {
		return nil, nil, err
	}

	calls, err := query.QueryCalls(tree, text, source)

	return symbols, calls, err
}

func (d *Document) extractLoads() {
	loads := []string{}
	for _, call := range d.Calls {
		if call.Name == builtin.LoadName {
			rawValue := call.Parameters[0].Value
			loads = append(loads, strings.Trim(rawValue, "\""))
		}
	}

	d.Loads = loads
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

func (d *Document) AddLoadedDoc(doc *Document) {
	d.DocLoads = append(d.DocLoads, doc)
}
