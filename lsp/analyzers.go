package lsp

import (
	"context"

	"github.com/tilt-dev/starlark-lsp/pkg/analysis"
)

func customAnalyzer(ctx context.Context) (*analysis.Analyzer, error) {
	// TODO: bsena; zero builtin for now
	var builtinDefPaths []string
	builtinAnalyzerOption := func() analysis.AnalyzerOption {
		return analysis.WithBuiltinPaths(builtinDefPaths)
	}

	opts := []analysis.AnalyzerOption{
		analysis.WithStarlarkBuiltins(),
		builtinAnalyzerOption(),
	}

	return analysis.NewAnalyzer(ctx, opts...)
}
