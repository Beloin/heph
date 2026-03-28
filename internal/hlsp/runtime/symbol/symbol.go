package symbol

type SymbolKind int

// Kind types
const (
	// Special symbol to define root scope
	RootKind SymbolKind = iota

	ClassKind
	FunctionKind
	VariableKind
	ValueKind
	StructKind
	// FieldKind is used for intermediate segments of dotted-path names (e.g. "heph" in "heph.pkg.dir").
	FieldKind
	// PrimitiveKind is used for built-in primitive type sentinels (int, str, bool, …).
	PrimitiveKind

	FunctionCallKind
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

	// Type points to the symbol representing this parameter's type.
	// nil means unknown. For primitives it points to a sentinel (e.g. PrimitiveInt).
	// For user-defined types it points to the class/struct symbol resolved at parse time.
	Type *Symbol
}

type Symbol struct {
	Name   string
	Source string

	FullyQualifiedName string

	Kind      SymbolKind
	Signature string

	// TODO: bsena; mayube instead of parameters we add then in symbol.Symbols bc we can know based on kind
	// that inerr symbols must be params
	// Actually we can use this as a "bypass" of symbols type, bc I am too lazy
	// to really use scope based queries in query.Calls
	Parameters []*Parameter

	// TODO: bsena; Maybe value can also be a Symbol, bc we can have a = b
	// Value is the current literal value for a variable
	Value     string
	DocString string

	Position Position

	SignaturePosition rawPosition

	Symbols []*Symbol

	// Type points to the symbol representing this symbol's type (nil if unknown).
	Type *Symbol
}

func (s *Symbol) Is(kind SymbolKind) bool {
	return s.Kind == kind
}

// TODO: bsena; rename after remove from common
// FindSymbolInSymbols searches only the top-level Symbols of parent for name.
func FindSymbolInSymbols(parent *Symbol, name string) (*Symbol, bool) {
	for _, s := range parent.Symbols {
		if s.Name == name {
			return s, true
		}
	}
	return nil, false
}

// Primitive type sentinels. These are the canonical *Symbol values for built-in types.
// Parameter.Type and Symbol.Type point to these for primitive types.
// For user-defined types they point to the class symbol resolved at parse time.
var (
	PrimitiveInt    = &Symbol{Name: "int", Kind: PrimitiveKind}
	PrimitiveFloat  = &Symbol{Name: "float", Kind: PrimitiveKind}
	PrimitiveBool   = &Symbol{Name: "bool", Kind: PrimitiveKind}
	PrimitiveNull   = &Symbol{Name: "Null", Kind: PrimitiveKind}
	PrimitiveString = &Symbol{Name: "str", Kind: PrimitiveKind}
	DictType        = &Symbol{Name: "dict", Kind: PrimitiveKind}
	ListType        = &Symbol{Name: "list", Kind: PrimitiveKind}
	ObjectType      = &Symbol{Name: "object", Kind: PrimitiveKind}
	UnknownType     = &Symbol{Name: "unknown", Kind: PrimitiveKind}
)
