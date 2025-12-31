package symbol

type SymbolKind int

// Kind types
const (
	ClassKind SymbolKind = iota
	FunctionKind
	VariableKind
	FieldKind
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
	Name string

	// TODO: bsena; Add this to validate full name of methods?
	// Or add reference to parent?
	// If its reference we need a subtree from a node of the tree.
	// Like a subtree for a custom node but search would be more painfully.

	// FullName references the compoosite name from class, function, method etc. names.
	// For example, `a.b.c.d` would be a full reference name for the symbol `d`.
	// Is a shortcut to ease queries
	FullName string

	Kind      SymbolKind
	Signature string

	// Value is the current literal value for a variable
	Value     string
	DocString string

	Position Position

	SignaturePosition rawPosition

	Symbols []*Symbol
}

func (s *Symbol) Is(kind SymbolKind) bool {
	return s.Kind == kind
}
