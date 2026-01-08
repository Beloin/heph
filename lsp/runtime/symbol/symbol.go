package symbol

import "strings"

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
	DefaultValue string
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

	// TODO: bsena; Creat a struct to have name and docstring extracted from "args" function docstring
	// also add type as string if it exists
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

// Args return symbol args if Symbol.Kind == FunctionKind
// TODO: bsena; Extract params from whitin query itself
func (s *Symbol) Args() []string {
	if !s.Is(FunctionKind) {
		return nil
	}

	return s.extractFunctionArgs()
}

// extractFunctionArgs extracts arguments from a function signature.
// Returns the arguments as a slice of strings and a boolean indicating if arguments were found.
func (s *Symbol) extractFunctionArgs() []string {
	signature, ok := strings.CutPrefix(s.Signature, s.Name)
	if !ok {
		return nil
	}

	// TODO: instead of cut, just do a slice
	signature, ok = strings.CutPrefix(signature, "(")
	if !ok {
		return nil
	}
	signature, ok = strings.CutSuffix(signature, ")")
	if !ok {
		return nil
	}

	if signature == "" {
		return nil
	}

	args := strings.Split(signature, ",")
	for i, arg := range args {
		args[i] = strings.TrimSpace(arg)
	}

	return args
}
