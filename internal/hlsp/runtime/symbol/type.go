package symbol

type Type struct {
	Name string
}

var (
	UnknownType = &Type{Name: "unknown"}

	PrimitiveInt    = &Type{Name: "int"}
	PrimitiveFloat  = &Type{Name: "float"}
	PrimitiveBool   = &Type{Name: "bool"}
	PrimitiveNull   = &Type{Name: "Null"}
	PrimitiveString = &Type{Name: "str"}

	DictType = &Type{Name: "dict"}
	ListType = &Type{Name: "list"}
)

// ResolveType resolves both Python type annotation names (e.g. "int", "str", "bool")
// and tree-sitter AST node kinds (e.g. "integer", "concatenated_string", "true", "dictionary").
func ResolveType(name string) *Type {
	switch name {
	case "int", "integer":
		return PrimitiveInt
	case "float":
		return PrimitiveFloat
	case "bool", "true", "false":
		return PrimitiveBool
	case "str", "string", "concatenated_string":
		return PrimitiveString
	case "None", "Null", "none":
		return PrimitiveNull
	case "dict", "dictionary":
		return DictType
	case "list":
		return ListType
	}

	return nil
}

// String returns the type name, or "unknown" for a nil receiver.
func (t *Type) String() string {
	if t == nil {
		return "unknown"
	}

	return t.Name
}

// IsKnown reports whether the type is non-nil and not the unknown sentinel.
func (t *Type) IsKnown() bool {
	return t != nil && t != UnknownType && t.Name != ""
}

// Is reports whether t is the same type as other (by pointer or name equality).
func (t *Type) Is(other *Type) bool {
	if t == nil || other == nil {
		return t == other
	}

	return t == other || t.Name == other.Name
}
