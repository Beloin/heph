package lang

import (
	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/hephbuild/heph/lsp/runtime/symbol"
	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var Logger = commonlog.GetLogger("completion")

func TextDocumentCompletionFuncWrapper(manager *runtime.Manager) protocol.TextDocumentCompletionFunc {
	return func(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
		// TODO: bsena; put the first from this uri
		var completionItems []protocol.CompletionItem

		// TODO: bsena; ~Read based in position so we can get classes' methods~
		// We actually are not going to do that, we will extract custom builtins based in hbuiltin

		// Need to implement:
		// 1. Inside function lookup its arguments
		// 2. Signature help

		// TODO: bsena: Also refactor this to just return completion in the end
		// And put the values in order

		// TODO: continue from here to look into function args and complete them
		// Whys is this not working?
		// If we have a function, we look first for its arguments

		// TODO: bsena; Prevent this when is function
		allPerKind := manager.AllLoadedSymbolsPerKind()

		if doc, found := manager.GetDocument(params.TextDocument.URI); found {
			byteOffset := params.Position.IndexIn(doc.TextString)
			symbolName := doc.ExtractCurrentSymbolName(uint(byteOffset))
			if s, found := manager.Query(symbolName); found {
				// If is function we can get args or other vars/functions as arguments
				if s.Is(symbol.FunctionKind) {
					args := s.Args()
					var completionItems []protocol.CompletionItem
					if args != nil {
						completionItems = createCompletionItemForArgs(args, s)
					}

					for _, symbol := range allPerKind.Variables {
						compItem := createCompletionItem(symbol)
						completionItems = append(completionItems, compItem)
					}

					for _, symbol := range allPerKind.Functions {
						compItem := createCompletionItem(symbol)
						completionItems = append(completionItems, compItem)
					}

					return completionItems, nil
				}
			}
		}

		for _, symbol := range allPerKind.AllSymbols {
			compItem := createCompletionItem(symbol)
			completionItems = append(completionItems, compItem)
		}

		return completionItems, nil
	}
}

func createCompletionItem(symbol *symbol.Symbol) protocol.CompletionItem {
	name := symbol.Name
	sig := symbol.Signature
	doc := symbol.DocString
	kind := MachineKindToCompletionKind(symbol.Kind)

	compItem := protocol.CompletionItem{
		Label:         name,
		InsertText:    &name,
		Kind:          &kind,
		Detail:        &sig,
		Documentation: doc,
	}
	return compItem
}

// TODO: this is ugly
func createCompletionItemForArgs(args []string, s *symbol.Symbol) []protocol.CompletionItem {
	completionItems := []protocol.CompletionItem{}
	for _, arg := range args {
		kind := protocol.CompletionItemKindField
		completionItems = append(completionItems, protocol.CompletionItem{
			Label:         arg,
			InsertText:    &arg,
			Kind:          &kind,
			Documentation: s.DocString,
		})
	}

	return completionItems
}
