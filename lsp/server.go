package lsp

import "context"

// TODO: bsena; NOW WE WILL IMPLEMENT FROM SCRATCH, USE https://github.com/tliron/glsp/ TO IMPLEMENT
// This will help to have custom capability and our own control over builtins

type LSPServer interface {
	Serve(ctx context.Context) error
	Close(ctx context.Context) error
}
