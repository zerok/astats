package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func generateQueryCmd() *Command {
	var topViewCount int
	var dbPath string
	var date string
	cmd := cobra.Command{
		Use:   "query METRIC",
		Short: "Query the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			if len(args) < 1 {
				return fmt.Errorf("Please specify a metric")
			}
			db, err := sql.Open("sqlite3", dbPath+"?mode=ro&cache=private")
			if err != nil {
				return err
			}
			defer db.Close()
			if date == "" {
				date = time.Now().UTC().Format("2006-01-02")
			}
			switch args[0] {
			case "top":
				res, err := db.QueryContext(ctx, "SELECT u.url, r.count FROM requests_per_day r JOIN urls u ON (u.id = r.url_id) WHERE r.date = ? ORDER BY r.count DESC LIMIT ?", date, topViewCount)
				if err != nil {
					return err
				}
				defer res.Close()
				for res.Next() {
					var u string
					var count int64
					if err := res.Scan(&u, &count); err != nil {
						return err
					}
					fmt.Printf("%5d %s\n", count, u)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&dbPath, "database-path", "astats.sqlite", "Path to the SQLite store")
	cmd.Flags().StringVar(&date, "date", "", "Date to be queried")
	cmd.Flags().IntVar(&topViewCount, "top", 10, "Show only the top n pages")
	return &Command{&cmd}
}
