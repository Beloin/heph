package lsp

import (
	"context"
	"embed"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/hephbuild/heph/hroot"
	"github.com/hephbuild/heph/vfssimple"
	"github.com/tilt-dev/starlark-lsp/pkg/analysis"
)

//go:embed builtin
var builtins embed.FS

// TODO: bsena; How to implement custom request?
// Search for custom protocol.Registration and jow starlark-lsp uses this

func customAnalyzer(ctx context.Context, root *hroot.State) (*analysis.Analyzer, error) {
	// TODO: bsena; This UGLY implementation is required since starlark-lsp does not loads embed.FS
	// and their API is weird
	builtinPath := root.Home.Join("heph-builtins")
	paths := []string{}

	err := fs.WalkDir(builtins, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		fullPath := filepath.Join(builtinPath.Abs(), path)
		dst, err := vfssimple.NewFile("file://" + fullPath)
		if err != nil {
			return err
		}
		defer dst.Close()

		src, err := builtins.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			return err
		}

		paths = append(paths, dst.Path())

		return nil
	})
	if err != nil {
		return nil, err
	}

	builtinAnalyzerOption := func() analysis.AnalyzerOption {
		return analysis.WithBuiltinPaths(paths)
	}

	opts := []analysis.AnalyzerOption{
		analysis.WithStarlarkBuiltins(),
		builtinAnalyzerOption(),
	}

	return analysis.NewAnalyzer(ctx, opts...)
}
