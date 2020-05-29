package main

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

type Command struct {
	*cobra.Command
}

func generateRootCmd() *Command {
	cmd := cobra.Command{
		Use: "astats",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	scanCmd := generateScanCmd()
	ingestCmd := generateIngestCmd()
	queryCmd := generateQueryCmd()
	cmd.AddCommand(scanCmd.Command)
	cmd.AddCommand(ingestCmd.Command)
	cmd.AddCommand(queryCmd.Command)
	return &Command{&cmd}
}

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	ctx := logger.WithContext(context.Background())
	cmd := generateRootCmd()
	if err := cmd.ExecuteContext(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Execution failed")
	}
}
