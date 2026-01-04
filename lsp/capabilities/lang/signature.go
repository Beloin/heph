package lang

import (
	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var SignatureLogger = commonlog.GetLogger("signature")

// TODO: bsena; Undestand where is this being called from
func TextDocumentSignatureHelpFuncWrapper(manager *runtime.Manager) protocol.TextDocumentSignatureHelpFunc {
	return func(context *glsp.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
		SignatureLogger.Noticef("Calling signature")
		if doc, found := manager.GetDocument(params.TextDocument.URI); found {
			// params.Context.IsRetrigger

			offSet := params.Position.IndexIn(doc.TextString)
			s, b := doc.ExtractCurrentSymbol(uint(offSet))

			SignatureLogger.Noticef("Siganture: found? %v -> %+v", b, s.Name)
		}

		isAvtive := protocol.UInteger(1)
		activeParam := protocol.UInteger(0)

		fakeSign := protocol.SignatureHelp{
			Signatures:      []protocol.SignatureInformation{
				{
					Label:           "hahahahahahahhahaha",
					Documentation:   "lorem ipsum dolor sit amet",
					Parameters:      []protocol.ParameterInformation{
						{
							Label: "ra\ndom text here asdasd asdsadasdad sadkasdlaksjd",
						},
					},
					ActiveParameter: &isAvtive,
				},
			},
			ActiveSignature: &isAvtive,
			ActiveParameter: &activeParam,
		}
		return &fakeSign, nil
	}
}
