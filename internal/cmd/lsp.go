package cmd

import (
	"errors"

	"github.com/hephbuild/heph/internal/engine"
	"github.com/hephbuild/heph/internal/hlsp"
	"github.com/spf13/cobra"
)

// heph lsp serve

// heph lsp s

func init() {
	lspCommand := &cobra.Command{
		Use:     "lsp",
		Aliases: []string{"lsp"},
		Short:   "Heph Language Server",
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("run heph Language Server with `serve`")
		},
	}

	cmdArgs := parseTargetMatcherArgs{cmdName: "serve"}
	servelspCmd := &cobra.Command{
		Use:               "serve",
		Short:             "Serve LSP",
		Aliases:           []string{"s"},
		ValidArgsFunction: cmdArgs.ValidArgsFunction(),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// localOpt := bootstrap.BootOpts{}
			// bs, err := bootstrap.BootBase(ctx, localOpt)
			// if err != nil {
			// 	return err
			// }

			root, err := engine.Root()
			if err != nil {
				return err
			}

			// TODO: bsena; Read yaml files to know drivers/plugins etc
			eg, err := newEngine(ctx, root)
			if err != nil {
				return err
			}

			// Use home to create log file
			server, err := hlsp.NewLSPServer(eg)
			if err != nil {
				return err
			}

			return server.Serve()
		},
	}

	lspCommand.AddCommand(servelspCmd)
	rootCmd.AddCommand(lspCommand)
}
