package document

import (
	"github.com/hephbuild/heph/lsp/runtime/builtin"
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

func NewDocument(tree *tree_sitter.Tree, rawText []byte) *Document {
	// pt := tree_sitter.NewLanguage(tree_sitter_python.Language())
	// q, _ := tree_sitter.NewQuery(pt, "")
	// cu := tree_sitter.NewQueryCursor()
	// matches := cu.Matches(q, tree.RootNode(), rawText)
	// first := matches.Next()
	// captures := first.Captures
	// capture := captures[0]
	//
	// newCaps := cu.Captures(q, tree.RootNode(), rawText)
	// first2, _ := newCaps.Next()
	// caps:= first2.Captures
	// caps[0].Index

	doc := &Document{Tree: tree, Text: rawText}
	doc.Symbols = builtin.ExtractSymbols(tree, rawText)

	// TODO: bsena; pre-fetch all functions, symbols from tree
	// Remember, trees are AST trees, so it should be something like this:
	//                         (*)
	//                       /    \
	//                     ( - )   (end)
	//                    /     \
	//                   1       2
	// doc.functions = query.Functions(doc, tree.RootNode())
	// doc.symbols = query.DocumentSymbols(doc)
	// doc.parseLoadStatements()
	// see builtin function, there I implemented a load function, so we should get that and put elsewhere
	return doc
}

// SwapTree atomic swaps current tree and return the closed old tree
func (d *Document) SwapTree(newT *tree_sitter.Tree) *tree_sitter.Tree {
	// TODO: bsena; make it atomic with sync.Mutex
	old := d.Tree
	old.Close()
	d.Tree = newT

	return old
}
