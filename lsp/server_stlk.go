package lsp

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/hephbuild/heph/hroot"
	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/server"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)


// stdioServer is a temporary implementation used as comparision to our own implementation
type stdioServer struct {
	skServer *server.Server

	client   protocol.Client
	jsonConn jsonrpc2.Conn

	// Force non-copy
	_ [0]sync.Mutex
}

// Serve blocks current goroutine accepting LSP requests
func (s *stdioServer) Serve(ctx context.Context) error {
	h := s.skServer.Handler(server.StandardMiddleware...)
	s.jsonConn.Go(ctx, h)

	select {
	case <-ctx.Done():
		_ = s.jsonConn.Close()
		return ctx.Err()
	case <-s.jsonConn.Done():
		if ctx.Err() == nil {
			if errors.Unwrap(s.jsonConn.Err()) != io.EOF {
				// only propagate connection error if context is still valid
				return s.jsonConn.Err()
			}
		}
	}

	return nil
}

func (s *stdioServer) Close(ctx context.Context) error {
	return s.skServer.Exit(ctx)
}

func NewStdioServer(ctx context.Context, root *hroot.State) (LSPServer, error) {
	stdio := struct {
		io.ReadCloser
		io.Writer
	}{
		os.Stdin,
		os.Stdout,
	}

	jsonConn := createJSONRPC(stdio)
	conn := createClient(ctx, jsonConn)

	docManager := document.NewDocumentManager()

	analyzer, err := customAnalyzer(ctx, root)
	if err != nil {
		return nil, err
	}

	_, cancel := context.WithCancel(ctx)
	s := server.NewServer(cancel, conn, docManager, analyzer)

	is := &stdioServer{
		skServer: s,
		jsonConn: jsonConn,
		client:   conn,
	}

	return is, nil
}

func createJSONRPC(conn io.ReadWriteCloser) jsonrpc2.Conn {
	stream := jsonrpc2.NewStream(conn)
	jsonConn := jsonrpc2.NewConn(stream)

	return jsonConn
}

func createClient(ctx context.Context, jsonConn jsonrpc2.Conn) protocol.Client {
	// I am not happy that it forces the zap logger usage
	logger := protocol.LoggerFromContext(ctx)
	return protocol.ClientDispatcher(jsonConn, logger.Named("notify"))
}
