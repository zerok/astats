package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/zerok/astats/pkg/accesslog"
)

type ViewCount struct {
	URI   string
	Count int64
}

type PageCountByCount []ViewCount

func (p PageCountByCount) Len() int {
	return len(p)
}
func (p PageCountByCount) Less(i, j int) bool {
	return p[i].Count < p[j].Count
}
func (p PageCountByCount) Swap(i, j int) {
	o := p[i]
	p[i] = p[j]
	p[j] = o
}

func isRelevantReferrer(ref string, ownDomain string) bool {
	if ref == "" {
		return false
	}
	if !strings.HasPrefix(ref, "http") {
		return false
	}
	skipDomains := []string{"duckduckgo.com", "fraidyc.at", "baidu.com", "t.co", "www.feedly.com", "feedly.com", "www.findyour.blog", "www.google", "m.baidu.com", ownDomain}
	for _, d := range skipDomains {
		if strings.HasPrefix(ref, "http://"+d) || strings.HasPrefix(ref, "https://"+d) {
			return false
		}
	}
	return true
}

func generateScanCmd() *Command {
	var todayOnly bool
	var filterContent string
	var topViewCount int
	var hide404 bool
	cmd := cobra.Command{
		Use:   "scan INPUTFILE",
		Short: "Load new log statements and show a summary",
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
			ny, nm, nd := time.Now().Date()
			range_start := time.Time{}
			range_end := time.Time{}
			daterange_start := time.Time{}
			daterange_end := time.Now()
			if todayOnly {
				daterange_start = time.Date(ny, nm, nd, 0, 0, 0, 0, time.Local)
			}
			views := make(map[string]int64)
			referrers := make(map[string]map[string]struct{})
			errs404 := make(map[string]int64)
			idx := -1
			for {
				line, err := lf.NextLine(ctx)
				if err != nil {
					if io.EOF != err {
						return err
					}
					break
				}
				if !((line.Time.After(daterange_start) || line.Time.Equal(daterange_start)) && (line.Time.Before(daterange_end) || line.Time.Equal(daterange_end))) {
					continue
				}
				if line.StatusCode == 404 {
					count := errs404[line.Request.URI]
					errs404[line.Request.URI] = count + 1
				}
				if filterContent != "" && line.ResponseHeaders.ContentType() != filterContent {
					continue
				}
				idx += 1
				if idx == 0 {
					range_start = line.Time
				}
				range_end = line.Time
				if line.StatusCode >= 200 && line.StatusCode < 300 {
					y, m, d := line.Time.Date()
					if y == ny && m == nm && d == nd {
						count := views[line.Request.URI]
						views[line.Request.URI] = count + 1
						ref := line.Request.Headers.Referrer()
						if isRelevantReferrer(ref, ownDomain) {
							refs, ok := referrers[line.Request.URI]
							if !ok {
								refs = make(map[string]struct{})
							}
							refs[line.Request.Headers.Referrer()] = struct{}{}
							referrers[line.Request.URI] = refs
						}
					}
				}
			}

			fmt.Printf("%s %s - %s\n\n", aurora.BrightWhite("Date range:").Bold(), range_start, range_end)

			counts := make([]ViewCount, 0, len(views))

			for uri, count := range views {
				counts = append(counts, ViewCount{URI: uri, Count: count})
			}

			fmt.Printf("%s\n%s\n", aurora.BrightWhite("Top posts:").Bold(), aurora.BrightWhite("----------").Bold())
			sort.Sort(sort.Reverse(PageCountByCount(counts)))
			for idx, v := range counts {
				if idx >= topViewCount {
					break
				}
				fmt.Printf("%5d %s\n", v.Count, v.URI)
			}

			if !hide404 {
				fmt.Printf("\n%s\n%s\n", aurora.BrightWhite("404 URLs:").Bold(), aurora.BrightWhite("---------").Bold())
				for u := range errs404 {
					fmt.Printf("%s\n", u)
				}
			}
			fmt.Printf("\n%s\n%s\n", aurora.BrightWhite("Referrers:").Bold(), aurora.BrightWhite("---------").Bold())
			for u, refs := range referrers {
				fmt.Printf("\n %s\n", u)
				for r := range refs {
					fmt.Printf("    %s\n", r)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&todayOnly, "today", false, "Set the date range to today-only")
	cmd.Flags().BoolVar(&hide404, "hide-404", false, "Hide 404 URLs")
	cmd.Flags().StringVar(&filterContent, "content-type", "", "Show only specific content types")
	cmd.Flags().IntVar(&topViewCount, "top", 10, "Show only the top n pages")
	return &Command{&cmd}
}
