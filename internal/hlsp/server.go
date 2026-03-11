package hlsp

import (
	"errors"
	"sync"

	"github.com/hephbuild/heph/internal/engine"
	"github.com/hephbuild/heph/internal/hfs"
	"github.com/hephbuild/heph/internal/hlsp/capabilities/lang"
	"github.com/hephbuild/heph/internal/hlsp/capabilities/lifecycle"
	docsync "github.com/hephbuild/heph/internal/hlsp/capabilities/sync"
	"github.com/hephbuild/heph/internal/hlsp/runtime"
	runtimedriver "github.com/hephbuild/heph/internal/hlsp/runtime/driver"

	"github.com/tliron/commonlog"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/tliron/glsp"
	"github.com/tliron/glsp/server"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

var ErrIsClosed = errors.New("server is closed")

// LSPServer implements the Heph Language Server Protocol.
// It uses STDIO for communication with client, so current STDIO should be controlled by the LSPServer.
type LSPServer interface {
	// Serve blocks until the server stops serving. Errors encountered during this process are returned.
	// Serve will be called only once for the lifetime of an LSPServer
	Serve() error
	Close() error
}

type hephLSP struct {
	protocolHandler *protocol.Handler
	server          *server.Server
	parser          *tree_sitter.Parser

	isClosed bool

	// Force non-copy
	_ [0]sync.Mutex
}

func NewLSPServer(e *engine.Engine) (LSPServer, error) {
	return newHephLSP(e, false)
}

func (h *hephLSP) Serve() error {
	if h.isClosed {
		return ErrIsClosed
	}

	return h.server.RunStdio()
}

func (h *hephLSP) Close() error {
	h.parser.Close()
	h.server.GetStdio().Close() //nolint
	h.isClosed = true

	return nil
}

func newHephLSP(engine *engine.Engine, debug bool) (*hephLSP, error) {
	err := configureLogs(&engine.Home, debug)
	if err != nil {
		return nil, err
	}

	lsp := &hephLSP{}

	parser := tree_sitter.NewParser()
	err = parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_python.Language()))
	if err != nil {
		return nil, err
	}

	registry := runtimedriver.NewRegistry(engine)
	manager, err := runtime.NewManager(parser, registry)
	if err != nil {
		return nil, err
	}

	handler := &protocol.Handler{
		// Lifecycle
		Initialize:  lsp.wrapInitialize(manager),
		Initialized: lsp.wrapInitialized(),
		Shutdown:    lsp.wrapShutdown(),
		SetTrace:    lsp.wrapSetTrace(),

		// Sync
		TextDocumentDidOpen:   docsync.TextDocumentDidOpenWrapper(manager),
		TextDocumentDidChange: docsync.TextDocumentDidChangeFuncWrapper(manager),

		// Lang features
		TextDocumentCompletion: lang.TextDocumentCompletionFuncWrapper(manager),
		TextDocumentHover:      lang.TextDocumentHoverFuncWrapper(manager),

		// Can be implemented Implement code lens to copy addr or give in virtual text a full path of target etc
		// TextDocumentCodeLens:                TextDocumentCodeLensFunc
		TextDocumentReferences:  lang.TextDocumentReferencesFuncWrapper(manager),
		TextDocumentDeclaration: lang.TextDocumentDeclarationFuncWrapper(manager),
		TextDocumentDefinition:  lang.TextDocumentDefinitionFuncWrapper(manager),

		WorkspaceDidRenameFiles: docsync.WorkspaceDidRenameFilesFunc(manager),
		WorkspaceDidDeleteFiles: docsync.WorkspaceDidDeleteFilesFunc(manager),
	}
	server := server.NewServer(handler, runtime.HephLanguage, debug)

	lsp.protocolHandler = handler
	lsp.server = server
	lsp.parser = parser

	return lsp, nil
}

func configureLogs(home *hfs.OS, debug bool) error {
	verbosity := 0
	if debug {
		verbosity = 2
	}

	f, err := hfs.Create(home, ".lsplogs")
	if err != nil {
		return err
	}
	defer f.Close()

	fullpath := home.Path(".lsplogs")
	// tliron/glsp forces us to use this weird logger
	commonlog.Configure(verbosity, &fullpath)

	return nil
}

func (h *hephLSP) wrapInitialize(manager *runtime.Manager) protocol.InitializeFunc {
	return func(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
		// Call lifecycle callback
		err := lifecycle.InitializeCallback(manager, context, params)
		if err != nil {
			return nil, err
		}

		capabilities := h.protocolHandler.CreateServerCapabilities()

		return protocol.InitializeResult{
			Capabilities: capabilities,
			ServerInfo: &protocol.InitializeResultServerInfo{
				Name:    runtime.HephLanguage,
				Version: &runtime.Version,
			},
		}, nil
	}
}

func (h *hephLSP) wrapInitialized() protocol.InitializedFunc {
	return func(context *glsp.Context, params *protocol.InitializedParams) error {
		return nil
	}
}

func (h *hephLSP) wrapShutdown() protocol.ShutdownFunc {
	return func(context *glsp.Context) error {
		protocol.SetTraceValue(protocol.TraceValueOff)
		return nil
	}
}

func (h *hephLSP) wrapSetTrace() protocol.SetTraceFunc {
	return func(context *glsp.Context, params *protocol.SetTraceParams) error {
		protocol.SetTraceValue(params.Value)
		return nil
	}
}
