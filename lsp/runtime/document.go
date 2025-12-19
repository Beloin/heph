package runtime

import (
	protocol "github.com/tliron/glsp/protocol_3_16"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

// TODO: bsena; implement something like this
// Think about how we will have multiple documents
type Document struct {
	Symbols []*protocol.DocumentSymbol
	Tree    *tree_sitter.Tree
}

func NewDocument(tree *tree_sitter.Tree) *Document {
	doc := &Document{Tree: tree}
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
