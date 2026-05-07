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
			if symbolName == "" || len(hierarchy) <= 2 { // Needs to have at least the symbol and root
				return nil, nil
			}

			var locations []protocol.Location
			upperScope := hierarchy[len(hierarchy)-2]
			locations = addScopeReferences(upperScope, symbolName)

			if upperScope.Kind == symbol.RootKind {
				// Search in Doc.LoadedBy
				doc.RangeIsLoadedBy(func(l *document.Load) {
					locations = append(locations, addScopeReferences(l.Doc.Root, symbolName)...)
				})
			}

			if len(locations) > 0 {
				return locations, nil
			}
		}

		return nil, nil
	}
}

func addScopeReferences(upperScope *symbol.Symbol, symName string) []protocol.Location {
	var locations []protocol.Location

	// Check if this symbol itself matches
	if upperScope.Name == symName {
		locations = append(locations, symbolLocation(upperScope.Source, upperScope))
	}

	// Check parameters for references
	for _, param := range upperScope.Parameters {
		if param.Value != nil {
			if param.Value.Name == symName || param.Value.Value == symName {
				locations = append(locations, symbolLocation(upperScope.Source, upperScope))
			}
		}
	}

	// Check symbols in this scope
	for _, sym := range upperScope.Symbols {
		// Recursively search nested scopes
		locations = append(locations, addScopeReferences(sym, symName)...)
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
