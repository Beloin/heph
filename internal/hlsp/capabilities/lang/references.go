package lang

import (
	"github.com/hephbuild/heph/internal/hlsp/runtime"
	"github.com/hephbuild/heph/internal/hlsp/runtime/document"
	"github.com/hephbuild/heph/internal/hlsp/runtime/symbol"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TextDocumentReferencesFuncWrapper(manager *runtime.Manager) protocol.TextDocumentReferencesFunc {
	return func(context *glsp.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
		if doc, found := manager.GetDocument(params.TextDocument.URI); found {
			pos := uint(params.Position.IndexIn(doc.TextString))

			_, symbolName, hierarchy := doc.SymbolHierarchyWithLocation(pos)
			if symbolName == "" || len(hierarchy) == 0 {
				return nil, nil
			}

			var locations []protocol.Location

			// Get the scope where the symbol is defined
			var definitionScope *symbol.Symbol
			for i := len(hierarchy) - 1; i >= 0; i-- {
				if hierarchy[i].Kind != symbol.FunctionCallKind && hierarchy[i].Kind != symbol.TargetCallKind {
					definitionScope = hierarchy[i]
					break
				}
			}

			if definitionScope == nil {
				return nil, nil
			}

			// Check if it's a root-level symbol (can be loaded by other docs)
			isRootLevel := len(hierarchy) <= 2 // [builtins, root, ...] or [root, ...]

			// Search for references within the definition scope
			locations = append(locations, findReferencesInScope(doc.FullPath, definitionScope, symbolName)...)

			// If root-level function, also search in documents that load this one
			if isRootLevel {
				doc.RangeIsLoadedBy(func(load *document.Load) {
					locations = append(locations, findReferencesInScope(load.Doc.FullPath, load.Doc.Root, symbolName)...)
				})
			}

			if len(locations) > 0 {
				return locations, nil
			}
		}

		return nil, nil
	}
}

func findReferencesInScope(fullPath string, scope *symbol.Symbol, symbolName string) []protocol.Location {
	var locations []protocol.Location

	// Check if this scope uses the symbol
	for _, sym := range scope.Symbols {
		// Check function/target calls
		if sym.Kind == symbol.FunctionCallKind || sym.Kind == symbol.TargetCallKind {
			if sym.Name == symbolName {
				locations = append(locations, symbolLocation(fullPath, sym))
			}
		}

		// Check parameters
		for _, param := range sym.Parameters {
			if param.Value != nil {
				if param.Value.Name == symbolName || param.Value.Value == symbolName {
					locations = append(locations, symbolLocation(fullPath, sym))
				}
			}
		}

		// Check symbol's value
		if sym.Value == symbolName {
			locations = append(locations, symbolLocation(fullPath, sym))
		}

		// Recursively search nested scopes (functions contain their own scope)
		if sym.Kind == symbol.FunctionKind {
			locations = append(locations, findReferencesInScope(fullPath, sym, symbolName)...)
		}
	}

	return locations
}

func symbolLocation(fullPath string, sym *symbol.Symbol) protocol.Location {
	return protocol.Location{
		URI: addProtocol(fullPath),
		Range: protocol.Range{
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
