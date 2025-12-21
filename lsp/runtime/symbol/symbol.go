package symbol

type SymbolKind int

// Kind types
const (
	FunctionKind SymbolKind = iota
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
	Name      string
	Kind      SymbolKind
	Signature string

	// Value is the current literal value for a variable, or doc string for functions
	Value string

	Position Position

	SignaturePosition rawPosition

	Symbols []*Symbol // TODO: bsena; inner field
}

func (s *Symbol) Is(kind SymbolKind) bool {
	return s.Kind == kind
}
