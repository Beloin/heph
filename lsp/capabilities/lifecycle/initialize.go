package lifecycle

import (
	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// TODO: When initialized, look from the root all BUILD files extracting all symbols for the manager
// Also put this into sync, looks better there.

var logger = commonlog.GetLogger("lifecycle")

// InitializeCallback is a callback that is called whithin InitializeFunc
func InitializeCallback(manager *runtime.Manager, context *glsp.Context, params *protocol.InitializeParams) error {
	root := ""
	if params.RootURI!= nil {
		root = *params.RootURI
	}
	logger.Noticef("InitializeCallback: folders=%v rootUri:%s", params.WorkspaceFolders, root)

	wkFolders := params.WorkspaceFolders
	if len(wkFolders) >= 1 {
		newVar := wkFolders[0]
		logger.Infof("Using workspace folder %q as %s", newVar.Name, newVar.URI)
		manager.WorkspaceFolder = newVar.URI
	}

	// TODO: bsena; look all BUILD files from the root and extract symbols for indexing (go routine?)

	return nil
}
