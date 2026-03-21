package lang

import (
	"fmt"
	"strings"

	"github.com/hephbuild/heph/internal/hlsp/runtime"
	"github.com/hephbuild/heph/internal/hlsp/runtime/builtin"
	"github.com/hephbuild/heph/internal/hlsp/runtime/query"
	"github.com/hephbuild/heph/internal/hlsp/runtime/symbol"
	"github.com/tliron/glsp"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TextDocumentHoverFuncWrapper(manager *runtime.Manager) protocol.TextDocumentHoverFunc {
	return func(glspContext *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
		doc, ok := manager.GetDocument(params.TextDocument.URI)
		if !ok {
			return &protocol.Hover{}, nil
		}
		bytePos := params.Position.IndexIn(doc.TextString)

		if literal := doc.ExtractCurrentStringLiteral(uint(bytePos)); literal != "" {
			return createLiteralHover(literal), nil
		}

		// TODO: bsena; Maybe create a chain-like validations? Looks cleaner

		// TODO: bsena; this does not get builtin like heph.pkg.addr

		// TODO: bsena; We need to find a way to get vars from inside a function
		// Maybe Look for WhereAmI before hierarchy? Or the other way around, look
		// whereami after hierarchy and then add docs if inside args
		symbolName := doc.ExtractCurrentSymbolName(uint(bytePos))

		// TODO: bsena; To make this works we actually need instead of current idenfitier name
		// we need to query "WordUnderCursor"
		// if symbolName == "" {
		// 	return nil, nil
		// }

		loc, symbolName, hierarchy := doc.SymbolHierarchyWithLocation(uint(bytePos))
		if symbolName == "" {
			return nil, nil
		}

		if loc == query.ArgsLocation && len(hierarchy) > 0 {
			s := hierarchy[len(hierarchy)-1]
			for _, p := range s.Parameters {
				if strings.Contains(p.Name, symbolName) {
					return createArgHover(s, p), nil
				}
			}
		}

		for i := len(hierarchy) - 1; i >= 0; i-- {
			current := hierarchy[i]
			if current.Kind == symbol.FunctionCallKind {
				continue
			}

			if current.Name == symbolName {
				return createHover(current), nil
			}
		}

		if symbolName == builtin.TargetName {
			if s := doc.QueryClosestTarget(uint(bytePos)); s != nil {
				return createHover(s), nil
			}
		}

		// If is an argument inside a function call we can get the function name and args information
		funName := doc.ExtractCurrentFunctionName(uint(bytePos))

		// Give target info from the first enclosing function call in document
		if funName == builtin.TargetName {
			if s := doc.QueryClosestTarget(uint(bytePos)); s != nil {
				for _, p := range s.Parameters {
					if strings.Contains(p.Name, symbolName) {
						return createArgHover(s, p), nil
					}
				}
			}
		}

		// TODO: bsena; maybe the chain can use this
		switch doc.WhereAmI(uint(bytePos)) {
		case query.BlockLocation:
			// TODO: bsena; THIS IS NOT WORKING
			if s, found := doc.Query(funName); found {
				if inner, found := symbol.FindSymbolInSymbols(s, symbolName); found {
					return createHover(inner), nil
				}
			}
		case query.ArgsLocation:
			if s, found := doc.Query(funName); found {
				for _, p := range s.Parameters {
					if strings.Contains(p.Name, symbolName) {
						return createArgHover(s, p), nil
					}
				}
			}

			// Actually I think we should do something smarter
			// maybe go through all nodes and extracting the parent symbol

		}

		if s, found := doc.Query(funName); found {
			for _, p := range s.Parameters {
				if strings.Contains(p.Name, symbolName) {
					return createArgHover(s, p), nil
				}
			}
		}

		// Query first for current document symbols
		if symbol, found := doc.Query(symbolName); found {
			return createHover(symbol), nil
		}

		if symbol, found := manager.Query(symbolName); found {
			return createHover(symbol), nil
		}

		return nil, nil
	}
}

func createHover(sym *symbol.Symbol) *protocol.Hover {
	var sb strings.Builder

	sb.WriteString(sym.Source)
	sb.WriteString("\n---\n")

	if sym.Is(symbol.VariableKind) {
		sig := sym.Name
		if sym.Type != nil {
			sig += ":" + sym.Type.Name
		}
		sig += " = " + sym.Value
		sb.WriteString(langDecorateMultiline(sig, runtime.HephLanguage))
	} else {
		sb.WriteString(langDecorateMultiline(sym.Signature, runtime.HephLanguage))
	}

	sb.WriteString("\n---\n")
	sb.WriteString(sym.DocString)

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: sb.String(),
		},
		Range: &protocol.Range{
			Start: protocol.Position{
				Line:      protocol.UInteger(sym.Position.RowStart),
				Character: protocol.UInteger(sym.Position.ColumnStart),
			},
			End: protocol.Position{
				Line:      protocol.UInteger(sym.Position.RowEnd),
				Character: protocol.UInteger(sym.Position.ColumnEnd),
			},
		},
	}
}

func createArgHover(fn *symbol.Symbol, param *symbol.Parameter) *protocol.Hover {
	var sb strings.Builder

	sb.WriteString(param.Name)

	if param.Type != nil {
		sb.WriteString(":" + param.Type.Name)
	}

	if param.Value != "" {
		sb.WriteString(" = " + param.Value)
	}

	sb.WriteString(" from " + langDecorate(fn.Name))
	if param.DocString != "" {
		sb.WriteString("\n---\n")
		sb.WriteString(param.DocString)
	}

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: sb.String(),
		},
	}
}

func createLiteralHover(literal string) *protocol.Hover {
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: parseDefinition(literal),
		},
	}
}

func langDecorateMultiline(text, lang string) string {
	return "```" + lang + "\n" + text + "\n```\n"
}

func langDecorate(text string) string {
	return "`" + text + "`"
}

func parseDefinition(literal string) string {
	if ok := strings.HasPrefix(literal, "//"); ok {
		return fmt.Sprintf("Heph Package: %q", literal)
	}

	return literal
}
