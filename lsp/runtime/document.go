package runtime

import (
	protocol "github.com/tliron/glsp/protocol_3_16"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type SymbolKind int

// Kind types
const (
	FunctionKind SymbolKind = iota
	VariableKind
)

type Position struct {
	RowStart    uint
	ColumnStart uint
	RowEnd      uint
	ColumnEnd   uint
}

type rawPosition struct {
	ByteStart uint
	ByteEnd   uint
}

type Symbol struct {
	Name      string
	Kind      SymbolKind
	Signature string

	// Value is the current literal value for a variable, or doc string for functions
	Value string

	Position Position

	signaturePosition rawPosition
}

func (s *Symbol) Is(kind SymbolKind) bool {
	return s.Kind == kind
}

// TODO: bsena; implement something like this
// Think about how we will have multiple documents
type Document struct {
	Symbols []*protocol.DocumentSymbol // TODO: bsena; we probalby will need to wrap this in our own type to handle docs etc
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
