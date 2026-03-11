package symbol

type SymbolKind int

// Kind types
const (
	ClassKind SymbolKind = iota
	FunctionKind
	VariableKind
	ValueKind
	StructKind
)

// TODO: bsena; Maybe we really actually need types
// Instead of only relying in Kind

// Calls types
const (
	FunctionCallKind = iota + 4
	// TODO: bsena; THIS IS NOT BEING USED
	TargetCallKind
)

type Position struct {
	ByteStart uint
	ByteEnd   uint

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
	Name string

	Value     string
	DocString string

	Type *Type
}

type Symbol struct {
	Name   string
	Source string

	FullyQualifiedName string

	// TODO: bsena; Maybe instead of Kind we can use only type?
	Kind      SymbolKind
	Signature string

	Parameters []*Parameter

	// Value is the current literal value for a variable
	Value     string
	DocString string

	Position Position

	SignaturePosition rawPosition

	Symbols []*Symbol

	Type Type
}

func (s *Symbol) Is(kind SymbolKind) bool {
	return s.Kind == kind
}
