package lang

import (
	"fmt"
	"strings"

	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/hephbuild/heph/lsp/runtime/symbol"
	"github.com/hephbuild/heph/specs"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TextDocumentHoverFuncWrapper(manager *runtime.Manager) protocol.TextDocumentHoverFunc {
	return func(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
		doc, ok := manager.GetDocument(params.TextDocument.URI)
		if !ok {
			return &protocol.Hover{}, nil
		}

		bytePos := params.Position.IndexIn(doc.TextString)
		// TODO: bsena; Make hover context-aware, look for:
		// - Target -> TO do this we would need to build the spec ourself, or use the DAG. See below
		// - Function
		// - Argument

		// TODO: bsena; also query for target //build/target:run
		// For now extract the targets from each document loaded into a target spec, so we can load it in runtime?
		// In the future is the best to have a Heph Server that runs and change at each file change, so we always have a fast
		// DAG available

		if literal := doc.ExtractCurrentStringLiteral(uint(bytePos)); literal != "" {
			return createLiteralHover(literal), nil
		}

		symbolName := doc.ExtractCurrentSymbolName(uint(bytePos))

		// If is an argument inside a function call we can get the function name and args information
		funName := doc.ExtractCurrentFunctionName(uint(bytePos))
		if s, found := manager.Query(funName); found {
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
		varSignature := langDecorateMultiline(sym.Name + " = " + sym.Value, runtime.HephLanguage)
		sb.WriteString(varSignature)
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
	paramText := param.Name
	if param.Type != "" {
		paramText += ":" + param.Type
	}

	if param.DefaultValue != "" {
		paramText += " = " + param.DefaultValue
	}

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: paramText + " from " + langDecorate(fn.Name),
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
	// TODO: bsena; not needed to cut //
	if s, ok := strings.CutPrefix(literal, "//"); ok {
		t, err := specs.ParseTargetAddr(s, literal)
		if err == nil {
			return fmt.Sprintf("%q Heph from %s", literal, t)
		}
	}

	return literal
}
