package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/zerok/astats/pkg/accesslog"
	"github.com/zerok/sqlitemigrate"
)

var migrations *sqlitemigrate.MigrationRegistry

func migrate(ctx context.Context, db *sql.DB) error {
	return migrations.Apply(ctx, db)
}

func init() {
	migrations = sqlitemigrate.NewRegistry()
	migrations.RegisterMigration([]string{
		`create table if not exists urls (
			id integer primary key autoincrement,
			url text not null unique
		 )`,

		`create table if not exists requests_per_day (
			url_id integer references urls(id) on delete cascade,
			date text not null,
			count int not null
		 )`,
		`create unique index requests_per_day_idx on requests_per_day(url_id, date)`,
	}, []string{})

	migrations.RegisterMigration([]string{
		`create table ingestion_states (
			id integer not null primary key autoincrement,
			log_timestamp text not null
		 )`,
	}, []string{})

	migrations.RegisterMigration([]string{
		`create table referrers (
			source_id integer not null references urls (id) on delete cascade,
			target_id integer not null references urls (id) on delete cascade
		 )`,
		`create unique index referrers_idx on referrers(source_id, target_id)`,
	}, []string{})
}

func generateIngestCmd() *Command {
	var dbPath string
	var lastReadTS int64
	cmd := cobra.Command{
		Use:   "ingest INPUTFILE",
		Short: "Load new log statements add add new entries to the database",
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
			db, err := sql.Open("sqlite3", dbPath)
			if err != nil {
				return err
			}
			defer db.Close()
			if err := migrate(ctx, db); err != nil {
				return err
			}
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				return err
			}
			lastReadTS, err = getLastReadTS(ctx, tx)
			if err != nil {
				tx.Rollback()
				return err
			}
			newViews := make(map[string]map[string]int64)
			newLines := make([]*accesslog.LogEntry, 0, 10)
			lastTS := lastReadTS
			for {
				entry, err := lf.NextLine(ctx)
				if io.EOF == err {
					break
				}
				if err != nil {
					tx.Rollback()
					return err
				}
				if lastReadTS != -1 && lastReadTS >= entry.Time.UnixNano() {
					continue
				}
				newLines = append(newLines, entry)
				lastTS = entry.Time.UnixNano()
			}

			for _, entry := range newLines {
				if entry.StatusCode != 200 || entry.ResponseHeaders.ContentType() != "text/html" {
					continue
				}
				date := entry.Time.UTC().Format("2006-01-02")
				viewCounts, ok := newViews[date]
				if !ok {
					viewCounts = make(map[string]int64)
				}
				viewCounts[entry.Request.URI] = viewCounts[entry.Request.URI] + 1
				newViews[date] = viewCounts
				ref := entry.Request.Headers.Referrer()
				if isRelevantReferrer(ref, ownDomain) {
					if err := addReferrer(ctx, tx, ref, entry.Request.URI); err != nil {
						tx.Rollback()
						return err
					}
				}
			}
			// Now let's add those to the database:
			for date, views := range newViews {
				for u, count := range views {
					urlID, err := getOrCreateURL(ctx, tx, u)
					if err != nil {
						tx.Rollback()
						return err
					}
					if err := incrementViewCount(ctx, tx, urlID, date, count); err != nil {
						tx.Rollback()
						return err
					}
				}
			}
			if err := updateInjestionState(ctx, tx, lastTS); err != nil {
				tx.Rollback()
				return err
			}
			fmt.Printf("Inserting %d new lines\n", len(newLines))
			return tx.Commit()
		},
	}
	cmd.Flags().StringVar(&dbPath, "database-path", "astats.sqlite", "Path to the SQLite store")
	cmd.Flags().Int64Var(&lastReadTS, "last-read-ts", -1, "Last read TS")
	return &Command{&cmd}
}

func updateInjestionState(ctx context.Context, tx *sql.Tx, ts int64) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO ingestion_states (log_timestamp) values (?)", ts)
	return err
}

func addReferrer(ctx context.Context, tx *sql.Tx, source string, target string) error {
	sourceID, err := getOrCreateURL(ctx, tx, source)
	if err != nil {
		return err
	}
	targetID, err := getOrCreateURL(ctx, tx, target)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO referrers (source_id, target_id) values (?, ?)", sourceID, targetID); err != nil {
		if e, ok := err.(sqlite3.Error); ok {
			if e.ExtendedCode == sqlite3.ErrConstraintUnique {
				return nil
			}
		}
		return err
	}
	return nil
}

func incrementViewCount(ctx context.Context, tx *sql.Tx, uid int64, date string, incr int64) error {
	var previousCount int64
	err := tx.QueryRowContext(ctx, "SELECT count FROM requests_per_day WHERE url_id = ? AND date = ?", uid, date).Scan(&previousCount)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	_, err = tx.ExecContext(ctx, "INSERT OR REPLACE INTO requests_per_day (url_id, date, count) VALUES (?, ?, ?)", uid, date, previousCount+incr)
	return err
}

func getOrCreateURL(ctx context.Context, tx *sql.Tx, u string) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, "SELECT id FROM urls WHERE url = ?", u).Scan(&id)
	if err == nil {
		return id, nil
	}
	if sql.ErrNoRows == err {
		res, err := tx.ExecContext(ctx, "INSERT INTO urls (url) values (?)", u)
		if err != nil {
			return -1, err
		}
		return res.LastInsertId()
	}
	return -1, err
}

func getLastReadTS(ctx context.Context, tx *sql.Tx) (int64, error) {
	var ts int64
	err := tx.QueryRowContext(ctx, "SELECT log_timestamp FROM ingestion_states ORDER BY id DESC LIMIT 1").Scan(&ts)
	if err == sql.ErrNoRows {
		return -1, nil
	}
	if err != nil {
		return ts, err
	}
	return ts, nil
}
