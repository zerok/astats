package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/zerok/astats/pkg/accesslog"
)

func generateIngestCmd() *Command {
	cmd := cobra.Command{
		Use:   "ingest INPUTFILE",
		Short: "Load new log statements into the datastore",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			if len(args) == 0 {
				return fmt.Errorf("no input file specified")
			}
			lf := accesslog.AccessLogFile{}
			fp, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer fp.Close()
			if err := lf.InitFromReader(fp); err != nil {
				return err
			}
			for {
				line, err := lf.NextLine(ctx)
				if err != nil {
					if io.EOF != err {
						return err
					}
					break
				}
				fmt.Println(line)
			}
			return nil
		},
	}
	var dbPath string
	cmd.Flags().StringVar(&dbPath, "database-path", "astats.sqlite", "Path to the SQLite store")
	return &Command{&cmd}
}
