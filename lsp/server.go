package lsp

import (
	"context"
	"errors"
	"sync"

	"github.com/hephbuild/heph/hroot"
	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/hephbuild/heph/vfssimple"

	"github.com/hephbuild/heph/lsp/capabilities/lang"
	"github.com/hephbuild/heph/lsp/capabilities/lifecycle"
	docsync "github.com/hephbuild/heph/lsp/capabilities/sync"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"

	"github.com/tliron/commonlog"
	_ "github.com/tliron/commonlog/simple"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
)

var ErrIsClosed = errors.New("server is closed")

// TODO: bsena; define better these interfaces

type LSPServer interface {
	// Serve blocks until the server stops serving. Errors encountered during this process are returned.
	// Serve will be called only once for the lifetime of an LSPServer
	Serve(ctx context.Context) error
	Close(ctx context.Context) error
}

type hephLSP struct {
	h       *protocol.Handler
	s       *server.Server
	p       *tree_sitter.Parser
	manager *runtime.Manager

	isClosed bool

	// Force non-copy
	_ [0]sync.Mutex
}

func NewHephServer(root *hroot.State) (LSPServer, error) {
	return newHephLSP(root, false)
}

func (h *hephLSP) Serve(ctx context.Context) error {
	if h.isClosed {
		return ErrIsClosed
	}

	return h.s.RunStdio()
}

func (h *hephLSP) Close(ctx context.Context) error {
	// TODO: bsena; How to Implement proper shutdown of lsp server in this method
	h.p.Close()
	h.s.GetStdio().Close() //nolint
	h.isClosed = true

	return nil
}

func newHephLSP(root *hroot.State, debug bool) (*hephLSP, error) {
	err := configureLogs(root, debug)
	if err != nil {
		return nil, err
	}

	lsp := &hephLSP{}

	// TODO: See if we can use Startlark tree-sitter
	// https://github.com/tree-sitter-grammars/tree-sitter-starlark
	parser := tree_sitter.NewParser()
	err = parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_python.Language()))
	if err != nil {
		return nil, err
	}

	// TODO: bsena; see if we can re-use parsers or not
	manager, err := runtime.NewManager(parser)
	if err != nil {
		return nil, err
	}

	// TODO: bsena; Add here custom capabilities and handler methods for our server
	handler := &protocol.Handler{
		// TODO: When initialized, look from the root all BUILD files extracting all symbols for the manager

		// Lifecycle
		Initialize:  lsp.wrapInitialize(manager),
		Initialized: lsp.wrapInitialized(),
		Shutdown:    lsp.wrapShutdown(),
		SetTrace:    lsp.wrapSetTrace(),

		// Sync
		TextDocumentDidOpen:   docsync.TextDocumentDidOpenWrapper(manager),
		TextDocumentDidChange: docsync.TextDocumentDidChangeFuncWrapper(manager),

		// Lang features
		// CompletionItemResolve:    lang.TextDocumentCompletionFuncWrapper(manager),
		TextDocumentCompletion:    lang.TextDocumentCompletionFuncWrapper(manager),
		TextDocumentHover:         lang.TextDocumentHoverFuncWrapper(manager),
		TextDocumentSignatureHelp: lang.TextDocumentSignatureHelpFuncWrapper(manager),

		// TextDocumentCodeLens:                TextDocumentCodeLensFunc // TODO: bsena; Implement code lens to copy addr?
		// TextDocumentReferences:  lang.TextDocumentReferencesFuncWrapper(manager),
		TextDocumentDeclaration: lang.TextDocumentDeclarationFuncWrapper(manager),
		TextDocumentDefinition:  lang.TextDocumentDefinitionFuncWrapper(manager),
	}
	server := server.NewServer(handler, runtime.HephLanguage, debug)

	lsp.h = handler
	lsp.s = server
	lsp.p = parser

	return lsp, nil
}

func configureLogs(root *hroot.State, debug bool) error {
	verbosity := 0
	if debug {
		verbosity = 2
	}

	// TODO: bsena; Find a way to prevent using this logger
	// and use default heph logger
	logpath := root.Home.Join("lsplogs")
	fullpath := logpath.Abs()
	dst, err := vfssimple.NewFile("file://" + fullpath)
	if err != nil {
		return err
	}
	defer dst.Close()

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

		capabilities := h.h.CreateServerCapabilities()

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
