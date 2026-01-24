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

		// If is function we can get args completion
		if doc, found := manager.GetDocument(params.TextDocument.URI); found {
			byteOffset := params.Position.IndexIn(doc.TextString)
			funName := doc.ExtractCurrentFunctionName(uint(byteOffset))

			if s, found := manager.Query(funName); found {
				if s.Is(symbol.FunctionKind) {
					args := s.Parameters
					if args != nil {
						completionItems = createCompletionItemForArgs(args, s)
					}
				}
			}

			// Append current doc symbols
			for _, s := range doc.Symbols {
				compItem := createCompletionItem(s)
				completionItems = append(completionItems, compItem)
			}

			// Complete symbols that are loaded by "load"
			for _, loadedDoc := range doc.DocLoads {
				for _, s := range loadedDoc.Symbols {
					compItem := createCompletionItem(s)
					completionItems = append(completionItems, compItem)
				}
			}

			return completionItems, nil
		}

		allPerKind := manager.AllLoadedSymbols()
		for _, symbol := range allPerKind {
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

func createCompletionItemForArgs(args []*symbol.Parameter, s *symbol.Symbol) []protocol.CompletionItem {
	completionItems := []protocol.CompletionItem{}
	for _, arg := range args {
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
