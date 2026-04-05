package query

import (
	"errors"
	"strconv"
	"strings"

	"github.com/hephbuild/heph/internal/hlsp/runtime/symbol"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

var ErrEmptyTreeError = errors.New("empty tree")

var lang = tree_sitter.NewLanguage(tree_sitter_python.Language())

func QueryAll(tree *tree_sitter.Tree, text []byte, source string) (*symbol.Symbol, error) {
	if tree.RootNode() == nil {
		return nil, ErrEmptyTreeError
	}

	rootScope := &symbol.Symbol{Kind: symbol.RootKind, Source: source}
	processScope(tree.RootNode(), text, rootScope)

	return rootScope, nil
}

func FilterSymbolsByKind(symbols []*symbol.Symbol, kind symbol.SymbolKind) []*symbol.Symbol {
	var result []*symbol.Symbol
	for _, s := range symbols {
		if s.Kind == kind {
			result = append(result, s)
		}
	}
	return result
}

func extractDocstring(node *tree_sitter.Node, text []byte) string {
	bodyNode := node.ChildByFieldName("body")
	if bodyNode == nil || bodyNode.ChildCount() == 0 {
		return ""
	}

	firstStmt := bodyNode.Child(0)
	if firstStmt == nil || firstStmt.Kind() != "expression_statement" {
		return ""
	}

	if firstStmt.ChildCount() == 0 {
		return ""
	}

	firstExpr := firstStmt.Child(0)
	if firstExpr == nil || firstExpr.Kind() != "string" {
		return ""
	}

	if firstExpr.ChildCount() < 2 {
		return ""
	}

	content := firstExpr.Child(1)
	if content == nil {
		return ""
	}

	doc := string(content.Utf8Text(text))
	return sanitizeComment(doc)
}

func sanitizeComment(cmmt string) string {
	if cmmt, ok := strings.CutPrefix(cmmt, "#"); ok {
		return processCommentLines(cmmt)
	}

	if cmmt, ok := strings.CutPrefix(cmmt, "\"\"\""); ok {
		cmmt, _ = strings.CutSuffix(cmmt, "\"\"\"")
		return processCommentLines(cmmt)
	}

	return processCommentLines(cmmt)
}

func processCommentLines(comment string) string {
	lines := strings.Split(comment, "\n")
	var processedLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			processedLines = append(processedLines, trimmed)
		}
	}

	return strings.Join(processedLines, "\n")
}

func parseArgsFromDocstring(docstring string, params []*symbol.Parameter) {
	lines := strings.Split(docstring, "\n")
	inArgs := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "Args:" {
			inArgs = true
			continue
		}
		if inArgs && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				namePart := strings.TrimSpace(parts[0])
				desc := strings.TrimSpace(parts[1])

				// If has type definition e.g param1 (int): My integer
				if idx := strings.Index(namePart, " ("); idx > 0 {
					paramName := namePart[:idx]
					for _, p := range params {
						if p.Name == paramName {
							p.DocString = desc
							break
						}
					}
				} else {
					for _, p := range params {
						if p.Name == namePart {
							p.DocString = desc
							break
						}
					}
				}
			}
		}
	}
}

func processScope(node *tree_sitter.Node, text []byte, parentSym *symbol.Symbol) {
	statements := getStatements(node)
	for _, stmt := range statements {
		switch stmt.Kind() {
		case "function_definition":
			processFunctionDefinition(stmt, text, parentSym)
		case "expression_statement":
			processExpressionStatement(stmt, text, parentSym)
		}
	}
}

// getStatements get first row statements
func getStatements(node *tree_sitter.Node) []*tree_sitter.Node {
	var statements []*tree_sitter.Node

	if node.Kind() == "module" || node.Kind() == "block" {
		count := node.ChildCount()
		for i := range count {
			child := node.Child(i)
			if child != nil {
				statements = append(statements, child)
			}
		}
	}

	return statements
}

func processFunctionDefinition(node *tree_sitter.Node, text []byte, parentSym *symbol.Symbol) {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return
	}

	nodeRange := node.Range()
	nameText := nameNode.Utf8Text(text)

	fn := &symbol.Symbol{
		Kind:               symbol.FunctionKind,
		Name:               nameText,
		FullyQualifiedName: nameText,
		Signature:          nameText + "()",
		Source:             parentSym.Source,
		Parent:             parentSym,
		Position: symbol.Position{
			RowStart:    nodeRange.StartPoint.Row,
			ColumnStart: nodeRange.StartPoint.Column,
			RowEnd:      nodeRange.EndPoint.Row,
			ColumnEnd:   nodeRange.EndPoint.Column,
			ByteStart:   nodeRange.StartByte,
			ByteEnd:     nodeRange.EndByte,
		},
	}

	fn.Parameters = extractParameters(node, text, parentSym)
	fn.DocString = extractDocstring(node, text)
	parseArgsFromDocstring(fn.DocString, fn.Parameters)

	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		processScope(bodyNode, text, fn)
	}

	parentSym.Symbols = append(parentSym.Symbols, fn)
}

func extractParameters(node *tree_sitter.Node, text []byte, scope *symbol.Symbol) []*symbol.Parameter {
	paramsNode := node.ChildByFieldName("parameters")
	if paramsNode == nil {
		return nil
	}

	var params []*symbol.Parameter
	count := paramsNode.NamedChildCount()
	for i := range count {
		paramNode := paramsNode.NamedChild(i)
		if paramNode == nil {
			continue
		}

		param := &symbol.Parameter{}
		kind := paramNode.Kind()
		switch kind {
		case "identifier":
			param.Name = paramNode.Utf8Text(text)
		case "default_parameter":
			if nameNode := paramNode.ChildByFieldName("name"); nameNode != nil {
				param.Name = nameNode.Utf8Text(text)
			} else if child := paramNode.Child(0); child != nil {
				param.Name = child.Utf8Text(text)
			}
			if valueNode := paramNode.ChildByFieldName("value"); valueNode != nil {
				param.Value = resolveValue(valueNode, text, scope)
			}
		case "typed_parameter":
			if nameNode := paramNode.ChildByFieldName("name"); nameNode != nil {
				param.Name = nameNode.Utf8Text(text)
			} else if child := paramNode.Child(0); child != nil && child.Kind() == "identifier" {
				param.Name = child.Utf8Text(text)
			}
			if typeNode := paramNode.ChildByFieldName("type"); typeNode != nil {
				param.Type = resolveType(typeNode.Utf8Text(text))
			}
		case "typed_default_parameter":
			if nameNode := paramNode.ChildByFieldName("name"); nameNode != nil {
				param.Name = nameNode.Utf8Text(text)
			} else if child := paramNode.Child(0); child != nil && child.Kind() == "identifier" {
				param.Name = child.Utf8Text(text)
			}

			if typeNode := paramNode.ChildByFieldName("type"); typeNode != nil {
				param.Type = resolveType(typeNode.Utf8Text(text))
			}

			if valueNode := paramNode.ChildByFieldName("value"); valueNode != nil {
				param.Value = resolveValue(valueNode, text, scope)
			}
		case "list_splat_pattern":
			if child := paramNode.Child(0); child != nil {
				param.Name = child.Utf8Text(text)
			}
		case "dictionary_splat_pattern":
			if child := paramNode.Child(0); child != nil {
				param.Name = child.Utf8Text(text)
			}
		}

		if param.Name != "" {
			params = append(params, param)
		}
	}

	return params
}

// resolveType resolves a type annotation string to a Symbol.
// For simple types (int, str), returns primitive sentinel.
// For complex types (List[str]), creates a temporary symbol.
// For user-defined types, searches scope chain.
func resolveType(typeStr string) *symbol.Symbol {
	// First check primitives
	switch typeStr {
	case "int":
		return symbol.PrimitiveInt
	case "float":
		return symbol.PrimitiveFloat
	case "bool":
		return symbol.PrimitiveBool
	case "str", "string":
		return symbol.PrimitiveString
	case "list":
		return symbol.ListType
	case "dict":
		return symbol.DictType
	}

	// We can include more types with generics etc
	// For now, create a placeholder symbol
	return &symbol.Symbol{Name: typeStr, Kind: symbol.PrimitiveKind}
}

// resolveValue resolves a default value or argument to a Symbol.
// For literals, creates a new Symbol with appropriate Kind.
// For references (identifiers, calls), resolves from scope.
func resolveValue(node *tree_sitter.Node, text []byte, scope *symbol.Symbol) *symbol.Symbol {
	symVal := node.Utf8Text(text)
	switch node.Kind() {
	case "string":
		return &symbol.Symbol{Kind: symbol.PrimitiveKind, Value: symVal, Type: symbol.PrimitiveString}
	case "integer":
		return &symbol.Symbol{Kind: symbol.PrimitiveKind, Value: symVal, Type: symbol.PrimitiveInt}
	case "float":
		return &symbol.Symbol{Kind: symbol.PrimitiveKind, Value: symVal, Type: symbol.PrimitiveFloat}
	case "true", "false":
		return &symbol.Symbol{Kind: symbol.PrimitiveKind, Value: symVal, Type: symbol.PrimitiveBool}
	case "list":
		return &symbol.Symbol{Kind: symbol.PrimitiveKind, Value: symVal, Type: symbol.ListType}
	case "dictionary":
		return &symbol.Symbol{Kind: symbol.PrimitiveKind, Value: symVal, Type: symbol.DictType}
	case "identifier":
		name := node.Utf8Text(text)
		if found := findSymbolInParentChain(scope, name); found != nil {
			return found
		}

		return &symbol.Symbol{Name: name, Kind: symbol.VariableKind, Type: symbol.UnknownType}
	case "call":
		return &symbol.Symbol{Kind: symbol.FunctionCallKind, Value: symVal}
	default:
		return &symbol.Symbol{Kind: symbol.ValueKind, Value: symVal, Type: symbol.UnknownType}
	}
}

func processExpressionStatement(stmt *tree_sitter.Node, text []byte, parentSym *symbol.Symbol) {
	childCount := stmt.ChildCount()
	for i := range childCount {
		child := stmt.Child(i)
		if child == nil {
			continue
		}

		switch child.Kind() {
		case "assignment":
			processAssignment(child, text, parentSym)
		case "call":
			processCall(child, text, parentSym)
		}
	}
}

func processAssignment(node *tree_sitter.Node, text []byte, parentSym *symbol.Symbol) {
	leftNode := node.ChildByFieldName("left")
	rightNode := node.ChildByFieldName("right")

	if leftNode == nil {
		return
	}

	nodeRange := node.Range()
	nameText := leftNode.Utf8Text(text)

	v := &symbol.Symbol{
		Kind:               symbol.VariableKind,
		Name:               nameText,
		FullyQualifiedName: nameText,
		Signature:          nameText,
		Source:             parentSym.Source,
		Parent:             parentSym,
		Position: symbol.Position{
			RowStart:    nodeRange.StartPoint.Row,
			ColumnStart: nodeRange.StartPoint.Column,
			RowEnd:      nodeRange.EndPoint.Row,
			ColumnEnd:   nodeRange.EndPoint.Column,
			ByteStart:   nodeRange.StartByte,
			ByteEnd:     nodeRange.EndByte,
		},
	}

	if rightNode != nil {
		v.Value = rightNode.Utf8Text(text)
		v.Type = inferTypeSymbol(rightNode, text, parentSym)
		extractCallsFromExpression(rightNode, text, v)
	}

	parentSym.Symbols = append(parentSym.Symbols, v)
}

func extractCallsFromExpression(node *tree_sitter.Node, text []byte, parentSym *symbol.Symbol) {
	if node == nil {
		return
	}

	if node.Kind() == "call" {
		processCall(node, text, parentSym)
		return
	}

	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			extractCallsFromExpression(child, text, parentSym)
		}
	}
}

func processCall(node *tree_sitter.Node, text []byte, parentSym *symbol.Symbol) {
	funcNode := node.ChildByFieldName("function")
	if funcNode == nil {
		return
	}

	nodeRange := node.Range()
	callName := getCallName(funcNode, text)

	call := &symbol.Symbol{
		Kind:               symbol.FunctionCallKind,
		Name:               callName,
		FullyQualifiedName: callName,
		Signature:          node.Utf8Text(text),
		Source:             parentSym.Source,
		Position: symbol.Position{
			RowStart:    nodeRange.StartPoint.Row,
			ColumnStart: nodeRange.StartPoint.Column,
			RowEnd:      nodeRange.EndPoint.Row,
			ColumnEnd:   nodeRange.EndPoint.Column,
			ByteStart:   nodeRange.StartByte,
			ByteEnd:     nodeRange.EndByte,
		},
	}

	argsNode := node.ChildByFieldName("arguments")
	if argsNode != nil {
		call.Parameters = extractCallArguments(argsNode, text, parentSym)
	}

	parentSym.Symbols = append(parentSym.Symbols, call)
}

func getCallName(node *tree_sitter.Node, text []byte) string {
	if node.Kind() == "identifier" {
		return node.Utf8Text(text)
	}

	if node.Kind() == "attribute" {
		var parts []string
		collectAttributeParts(node, text, &parts)
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}

	return ""
}

func collectAttributeParts(node *tree_sitter.Node, text []byte, parts *[]string) {
	if node.Kind() == "identifier" {
		*parts = append(*parts, node.Utf8Text(text))
		return
	}

	if node.Kind() == "attribute" {
		if objNode := node.ChildByFieldName("object"); objNode != nil {
			collectAttributeParts(objNode, text, parts)
		}
		if attrNode := node.ChildByFieldName("attribute"); attrNode != nil {
			*parts = append(*parts, attrNode.Utf8Text(text))
		}
	}
}

func extractCallArguments(node *tree_sitter.Node, text []byte, scope *symbol.Symbol) []*symbol.Parameter {
	count := node.NamedChildCount()
	args := make([]*symbol.Parameter, 0, count)
	for i := range count {
		argNode := node.NamedChild(i)
		if argNode == nil {
			continue
		}

		arg := &symbol.Parameter{}

		if argNode.Kind() == "keyword_argument" {
			if nameNode := argNode.ChildByFieldName("name"); nameNode != nil {
				arg.Name = nameNode.Utf8Text(text)
			}

			if valueNode := argNode.ChildByFieldName("value"); valueNode != nil {
				arg.Value = resolveValue(valueNode, text, scope)
			}
		} else {
			arg.Name = strconv.Itoa(int(i))
			arg.Value = resolveValue(argNode, text, scope)
		}

		args = append(args, arg)
	}

	return args
}

func inferTypeSymbol(node *tree_sitter.Node, text []byte, sym *symbol.Symbol) *symbol.Symbol {
	switch node.Kind() {
	case "string":
		return symbol.PrimitiveString
	case "integer":
		return symbol.PrimitiveInt
	case "float":
		return symbol.PrimitiveFloat
	case "true", "false":
		return symbol.PrimitiveBool
	case "list":
		return symbol.ListType
	case "dictionary":
		return symbol.DictType
	case "identifier":
		name := node.Utf8Text(text)
		if found := findSymbolInParentChain(sym, name); found != nil {
			return found
		}
		return &symbol.Symbol{Name: name, Kind: symbol.VariableKind}
	case "call":
		return symbol.UnknownType
	case "attribute":
		return &symbol.Symbol{Name: getCallName(node, text), Kind: symbol.FieldKind}
	default:
		return symbol.UnknownType
	}
}

func findSymbolInParentChain(sym *symbol.Symbol, name string) *symbol.Symbol {
	if sym == nil {
		return nil
	}

	for _, s := range sym.Symbols {
		if s.Name == name {
			return s
		}
	}

	return findSymbolInParentChain(sym.Parent, name)
}

func getFunctionNameNodeIfExists(node *tree_sitter.Node) *tree_sitter.Node {
	for funNode := node; funNode.Parent() != nil; funNode = funNode.Parent() {
		if funNode.Kind() != "function_definition" {
			continue
		}

		nameNode := funNode.ChildByFieldName("name")
		if nameNode == nil {
			break
		}

		return nameNode
	}

	return nil
}
