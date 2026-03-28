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
			if symbolName == "" {
				return nil, nil
			}

			var locations []protocol.Location
			for i := len(hierarchy) - 1; i >= 0; i-- {
				level := hierarchy[i]
				for _, sym := range level.Symbols {
					if sym.Name == symbolName {
						locations = append(locations, symbolLocation(doc.FullPath, sym))
					}
				}
				if level.Kind == symbol.RootKind {
					break
				}
			}

			doc.RangeIsLoadedBy(func(load *document.Load) {
				locations = append(locations, findReferencesInDoc(load.Doc, symbolName)...)
			})

			if len(locations) > 0 {
				return locations, nil
			}
		}

		return nil, nil
	}
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

// TODO: We will probably later need a better usage of symbols,
// having a symbol oriented query instead of doc based queries
func findReferencesInDoc(doc *document.Document, symbolName string) []protocol.Location {
	var locations []protocol.Location

	// TODO: bsena; Query for all things, but make sure to respect scopes

	for _, sym := range doc.QueryCalls(symbolName) {
		locations = append(locations, symbolLocation(doc.FullPath, sym))
	}

	return locations
}
