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
	return func(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
		var completionItems []protocol.CompletionItem

		for word, emoji := range EmojiMapper {
			emojiCopy := emoji // Create a copy of emoji
			completionItems = append(completionItems, protocol.CompletionItem{
				Label:      word,
				Detail:     &emojiCopy,
				InsertText: &emojiCopy,
			})
		}

		return completionItems, nil
	}
}
