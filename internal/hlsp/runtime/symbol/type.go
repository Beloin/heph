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
)

func ResolveType(name string) *Type {
	switch name {
	case "int":
		return PrimitiveInt
	case "float":
		return PrimitiveFloat
	case "bool":
		return PrimitiveBool
	case "str":
		return PrimitiveString
	case "None", "Null":
		return PrimitiveNull
	}
	return nil
}
