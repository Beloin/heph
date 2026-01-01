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
	Name   string
	Source string

	// TODO: bsena; Add this to validate full name of methods?
	// Or add reference to parent?
	// If its reference we need a subtree from a node of the tree.
	// Like a subtree for a custom node but search would be more painfully.

	// FullyQualifiedName references the compoosite name from class, function, method etc. names.
	// For example, `a.b.c.d` would be a full reference name for the symbol `d`.
	// Is a shortcut to ease queries
	FullyQualifiedName string

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

// Args return symbol args if Symbol.Kind == FunctionKind
func (s *Symbol) Args() []string {
	if !s.Is(FunctionKind) {
		return nil
	}

	return extractFunctionArgs(s.Signature)
}

// extractFunctionArgs extracts arguments from a function signature.
// Returns the arguments as a slice of strings and a boolean indicating if arguments were found.
func extractFunctionArgs(signature string) []string {
	signature, ok := strings.CutPrefix(signature, "(")
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
