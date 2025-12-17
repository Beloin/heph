package runtime

import (
	protocol "github.com/tliron/glsp/protocol_3_16"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

// TODO: bsena; Have something like this to have all current language symbols, functions methods, types etc
// sitter.Tree
// type Builtins struct {
// 	Functions map[string]query.Signature
// 	Symbols   []query.Symbol
// 	Types     map[string]query.Type
// 	Methods   map[string]query.Signature
// 	Members   []query.Symbol
// }

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
	// TODO: bsena; make it atomic
	old := d.Tree
	old.Close()
	d.Tree = newT

	return old
}

//	type TypeRef struct {
//		Name        string
//		Description string
//	}
type Symbol struct {
	Name        string
	Description string
	Kind        protocol.DocumentSymbol
}

//
// type Function struct {
// 	Name        string
// 	Signature   string
// 	Description string
// 	Symbols     []Symbol
// }
//
// type Method struct {
// 	Name        string
// 	Receiver    string
// 	Signature   string
// 	Description string
// 	Symbols     []Symbol
// }
//
