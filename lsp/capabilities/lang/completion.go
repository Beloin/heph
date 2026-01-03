package lang

import (
	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/hephbuild/heph/lsp/runtime/symbol"
	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var Logger = commonlog.GetLogger("completion")

// TODO: bsena; Implement protocol.CompletionItemResolveFunc

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
			Logger.Noticef("Found Doc: %s", doc.Name)

			byteOffset := params.Position.IndexIn(doc.TextString)
			symbolName := doc.ExtractCurrentSymbolName(uint(byteOffset))
			Logger.Noticef("Offset: %d & SymbolName: %s", byteOffset, symbolName)
			
			// TODO: Still not working, Symbol and Funname are empty
			funName := doc.ExtractCurrentFunctionName(uint(byteOffset))
			Logger.Noticef("Offset: %d & FunName: %s", byteOffset, funName)

			// TODO: looks like it is trapped inside function () -> How to go outside? Request parent node?
			// But we will actually need to have parameters splicit in symbol, since we can have a lot of parans werdly sparsed
			if s, found := manager.Query(funName); found {
				Logger.Noticef("Symbol Found: %s", s.Signature)

				// If is function we can get args or other vars/functions as arguments
				if s.Is(symbol.FunctionKind) {
					args := s.Parameters
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
func createCompletionItemForArgs(args []*symbol.Parameter, s *symbol.Symbol) []protocol.CompletionItem {
	completionItems := []protocol.CompletionItem{}
	for _, arg := range args {
		kind := protocol.CompletionItemKindField
		completionItems = append(completionItems, protocol.CompletionItem{
			Label:         arg.Name,
			InsertText:    &arg.Name,
			Kind:          &kind,
			Documentation: s.DocString,
		})
	}

	return completionItems
}
