package lang

import (
	"github.com/hephbuild/heph/internal/hlsp/runtime"
	"github.com/hephbuild/heph/internal/hlsp/runtime/query"
	"github.com/hephbuild/heph/internal/hlsp/runtime/symbol"
	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var Logger = commonlog.GetLogger("completion")

func TextDocumentCompletionFuncWrapper(manager *runtime.Manager) protocol.TextDocumentCompletionFunc {
	return func(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
		var completionItems []protocol.CompletionItem

		if doc, found := manager.GetDocument(params.TextDocument.URI); found {
			byteOffset := params.Position.IndexIn(doc.TextString)
			offset := uint(byteOffset)

			loc, symbolName, hierarchy := doc.SymbolHierarchyWithLocation(offset)

			// TODO: bsena; when is target(driver="") -> show available drivers

			if loc == query.ArgsLocation {
				// Find the closest function/target call in the hierarchy
				for i := len(hierarchy) - 1; i >= 0; i-- {
					current := hierarchy[i]
					if current.Kind == symbol.FunctionCallKind {
						continue
					}

					// Check nested symbols within this hierarchy level
					for _, sym := range current.Symbols {
						if sym.Name == symbolName {
							completionItems = createCompletionItemForArgs(sym)
							break
						}
					}
				}
			}

			// Add completion items for each symbol in hierarchy
			// Ignoring function calls
			for i := len(hierarchy) - 1; i >= 0; i-- {
				current := hierarchy[i]
				if len(current.Symbols) > 0 {
					completionItems = append(completionItems, createCompletionItemsForSymbols(current.Symbols)...)
				}
			}
		}

		return completionItems, nil
	}
}

func createCompletionItem(symbol *symbol.Symbol) protocol.CompletionItem {
	name := symbol.Name
	sig := symbol.Signature
	doc := symbol.DocString
	kind := SymbolKindToCompletionKind(symbol.Kind)

	compItem := protocol.CompletionItem{
		Label:         name,
		InsertText:    &name,
		Kind:          &kind,
		Detail:        &sig,
		Documentation: doc,
	}
	return compItem
}

func createCompletionItemsForSymbols(syms []*symbol.Symbol) []protocol.CompletionItem {
	items := []protocol.CompletionItem{}
	for _, s := range syms {
		if s.Kind == symbol.FunctionCallKind ||
			s.Kind == symbol.TargetCallKind ||
			s.Kind == symbol.RootKind {
			continue
		}

		items = append(items, createCompletionItem(s))
	}
	return items
}

func createCompletionItemForArgs(s *symbol.Symbol) []protocol.CompletionItem {
	completionItems := []protocol.CompletionItem{}
	for _, arg := range s.Parameters {
		kind := protocol.CompletionItemKindVariable

		label := arg.Name + "="
		completionItems = append(completionItems, protocol.CompletionItem{
			Label:         label,
			InsertText:    &label,
			Kind:          &kind,
			Documentation: s.DocString,
		})
	}

	return completionItems
}
