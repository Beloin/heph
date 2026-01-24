package symbol

type SymbolKind int

// Kind types
const (
	ClassKind SymbolKind = iota
	FunctionKind
	VariableKind
	FieldKind
)

// Calls types
const (
	FunctionCallKind = iota + 4
	TargetCallKind
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

type Parameter struct {
	Name         string
	Type         string
	Value string
	DocString    string
}

type Symbol struct {
	Name   string
	Source string

	// TODO: bsena; Test if query.ExtractFunction can read FullyQualifiedName like heph.path.cwd
	// FullyQualifiedName references the compoosite name from class, function, method etc. names.
	// For example, `a.b.c.d` would be a full reference name for the symbol `d`.
	// Is a shortcut to ease queries
	FullyQualifiedName string

	Kind      SymbolKind
	Signature string

	Parameters []*Parameter

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
