package lang

import (
	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// protocol.TextDocumentCompletionFunc

var EmojiMapper = map[string]string{
	"happy":      "😀",
	"sad":        "😢",
	"angry":      "😠",
	"confused":   "😕",
	"excited":    "😆",
	"love":       "😍",
	"laughing":   "😂",
	"crying":     "😭",
	"sleepy":     "😴",
	"surprised":  "😮",
	"sick":       "🤒",
	"cool":       "😎",
	"nerd":       "🤓",
	"worried":    "😟",
	"scared":     "😨",
	"silly":      "🤪",
	"shocked":    "😱",
	"sunglasses": "😎",
	"tongue":     "😛",
	"thinking":   "🤔",
}

func TextDocumentCompletionFuncWrapper(manager *runtime.Manager) protocol.TextDocumentCompletionFunc {
	// TODO: bsena; Add builins as docs in manager doc map like heph://builtin
	bts := runtime.ParseBuiltins(manager.Parser)
	return func(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
		// TODO: search in manager
		// params.TextDocument.URI
		var completionItems []protocol.CompletionItem

		for _, symbol := range bts {
			detail := *symbol.Detail
			name := symbol.Name
			// TODO: bsena; Parse kind properly based on symbol.Kind
			kind := protocol.CompletionItemKindConstructor

			// TODO: bsena; How to add parameters etc etc?
			completionItems = append(completionItems, protocol.CompletionItem{
				Label:      name,
				InsertText: &name,
				Kind:       &kind,
				Detail:     &name, // TODO: bsena; maybe details is the signature of the function, or the value itself of the thing
				Documentation: &detail,
			})
		}

		return completionItems, nil
	}
}
