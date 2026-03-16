package query

import (
	"github.com/hephbuild/heph/internal/hlsp/runtime/symbol"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

// TODO: bsena; use doc.QueryType
// ResolveType maps both Python type annotation names (e.g. "int", "str", "bool")
// and tree-sitter AST node kinds (e.g. "integer", "concatenated_string", "true", "dictionary")
// to the corresponding primitive symbol sentinel. Returns nil for unrecognised names.
func ResolveType(name string) *symbol.Symbol {
	switch name {
	case "int", "integer":
		return symbol.PrimitiveInt
	case "float":
		return symbol.PrimitiveFloat
	case "bool", "true", "false":
		return symbol.PrimitiveBool
	case "str", "string", "concatenated_string":
		return symbol.PrimitiveString
	case "None", "Null", "none":
		return symbol.PrimitiveNull
	case "dict", "dictionary":
		return symbol.DictType
	case "list":
		return symbol.ListType
	}
	return nil
}

// ExtractCurrentSymbol from root that is an identifier
func ExtractCurrentSymbol(root *tree_sitter.Node, source []byte, byteOffSet uint) string {
	for root != nil && root.Kind() != "identifier" {
		root = root.FirstChildForByte(byteOffSet)
	}

	if root == nil {
		return ""
	}

	return root.Utf8Text(source)
}

// ExtractCurrentStringLiteral finds closes string content
func ExtractCurrentStringLiteral(root *tree_sitter.Node, source []byte, byteOffSet uint) string {
	for root != nil && root.Kind() != "string_content" {
		root = root.FirstChildForByte(byteOffSet)
	}

	if root == nil {
		return ""
	}

	return root.Utf8Text(source)
}

type CodeLocation int

const (
	RootLocation  CodeLocation = iota // top-level, not inside any function or call
	BlockLocation                     // inside a function definition body
	ArgsLocation                      // inside a function call's argument list
)

// WhereAmI descends to the deepest node at byteOffset, then walks up the tree
// to determine context. First match going up wins:
//   - function_definition → ArgsLocation (cursor is on a parameter in a definition)
//   - call               → ArgsLocation (cursor is inside a call's argument list)
//   - block              → BlockLocation (cursor is inside a function body)
//   - otherwise          → RootLocation
func WhereAmI(root *tree_sitter.Node, source []byte, byteOffSet uint) CodeLocation {
	// Descend to the deepest node at the offset.
	for child := root.FirstChildForByte(byteOffSet); child != nil; child = child.FirstChildForByte(byteOffSet) {
		root = child
	}

	// Walk up to find context — first match wins.
	for node := root; node != nil; node = node.Parent() {
		switch node.Kind() {
		case "function_definition", "call":
			return ArgsLocation
		case "block":
			return BlockLocation
		}
	}

	return RootLocation
}

// SymbolHierarchy converts a node-position hierarchy to a Symbol hierarchy.
// It descends the AST following byteOffset, and at each identifier or
// function_definition node looks up the name in rootSyms. If found, the symbol
// is appended to the result and the search narrows to that symbol's sub-symbols
// for the next level.
func SymbolHierarchy(root *tree_sitter.Node, source []byte, byteOffSet uint, rootSyms []*symbol.Symbol) []*symbol.Symbol {
	var hierarchySymbols []*symbol.Symbol
	for child := root.FirstChildForByte(byteOffSet); child != nil; child = child.FirstChildForByte(byteOffSet) {
		root = child

		name := ""
		if child.Kind() == "identifier" {
			name = child.Utf8Text(source)
		} else if child.Kind() == "function_definition" {
			if nameNode := child.ChildByFieldName("name"); nameNode != nil {
				name = nameNode.Utf8Text(source)
			}
		}

		if name == "" {
			continue
		}

		var s *symbol.Symbol
		for _, sym := range rootSyms {
			if sym.Name != name {
				continue
			}
			if sym.Position.ByteStart > byteOffSet {
				continue
			}
			if s == nil || sym.Position.ByteStart > s.Position.ByteStart {
				s = sym
			}
		}
		if s == nil {
			break
		}

		hierarchySymbols = append(hierarchySymbols, s)
		rootSyms = s.Symbols
	}

	return hierarchySymbols
}

// ExtractFunctionNameFromOffset extracts closest current function name whether its a call or definition
func ExtractFunctionNameFromOffset(root *tree_sitter.Node, source []byte, byteOffSet uint) string {
	childLookup := ""
	for root != nil {
		switch root.Kind() {
		case "function_definition":
			childLookup = "name"
		case "call":
			childLookup = "function"
		}

		if childLookup != "" {
			break
		}

		root = root.FirstChildForByte(byteOffSet)
	}

	if root == nil {
		return ""
	}

	root = root.ChildByFieldName(childLookup)

	if root == nil {
		return ""
	}

	return root.Utf8Text(source)
}
