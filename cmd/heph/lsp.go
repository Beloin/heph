package main

import (
	"fmt"

	"github.com/hephbuild/heph/bootstrap"
	"github.com/hephbuild/heph/lsp"
	"github.com/spf13/cobra"
)

func init() {
	lspCommand.AddCommand(servelspCmd)
}

var lspCommand = &cobra.Command{
	Use:     "lsp",
	Aliases: []string{"lsp"},
	Short:   "Heph Language Server",
	Args:    cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("run heph Language Server with `serve`")
	},
}

var servelspCmd = &cobra.Command{
	Use:               "serve",
	Short:             "Serve LSP",
	Aliases:           []string{"s"},
	ValidArgsFunction: ValidArgsFunctionTargets,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: bsena; Probably we need to add options such as "--debug" and "--verbose"
		// TODO: bsena; To implement more about the starlark, see --builtin-paths

		// TODO: bsena; glsp stoles our ^C, only close with ^D, fix it
		ctx := cmd.Context()

		// TODO: bsena; This KILLS MY LSP SERVER DONT DO THIS
		// find another way to get ROOT

		localOpt := bootstrap.BootOpts{}
		bs, err := bootstrap.BootBase(ctx, localOpt)
		if err != nil {
			return err
		}

		server, err := lsp.NewHephServer(bs.Root)
		if err != nil {
			return err
		}

		return server.Serve(ctx)
	},
}
