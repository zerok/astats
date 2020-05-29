package accesslog

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"
)

type LineFormat struct {
	Format string
}

type AccessLogFile struct {
	TimestampFormat string
	Path            string
	r               *bufio.Reader
}

type LogEntryResponseHeaders struct {
	RawContentTypes []string `json:"Content-Type"`
	RawUserAgents   []string `json:"User-Agent"`
}

func (h *LogEntryResponseHeaders) ContentType() string {
	if len(h.RawContentTypes) > 0 {
		elems := strings.Split(h.RawContentTypes[0], ";")
		if len(elems) > 0 {
			return elems[0]
		}
		return ""
	}
	return ""
}

type LogEntryRequestHeaders struct {
	RawReferrers []string `json:"Referer"`
}

func (h *LogEntryRequestHeaders) Referrer() string {
	if len(h.RawReferrers) > 0 {
		return h.RawReferrers[0]
	}
	return ""
}

type LogEntryRequest struct {
	Method     string                 `json:"method"`
	URI        string                 `json:"uri"`
	RemoteAddr string                 `json:"remote_addr"`
	Host       string                 `json:"host"`
	Headers    LogEntryRequestHeaders `json:"headers"`
}

type LogEntry struct {
	Time            time.Time
	Timestamp       float64                 `json:"ts"`
	Size            int64                   `json:"size"`
	Duration        float64                 `json:"duration"`
	StatusCode      int                     `json:"status"`
	Request         LogEntryRequest         `json:"request"`
	ResponseHeaders LogEntryResponseHeaders `json:"resp_headers"`
}

func (lf *AccessLogFile) InitFromReader(r io.Reader) error {
	lf.r = bufio.NewReader(r)
	return nil
}

func (lf *AccessLogFile) NextLine(ctx context.Context) (*LogEntry, error) {
	line, err := lf.r.ReadString('\n')
	if err != nil {
		if line == "" {
			return nil, err
		}
	}
	result := &LogEntry{}
	if err := json.Unmarshal([]byte(line), result); err != nil {
		return nil, err
	}
	nanos := int64((result.Timestamp - float64(int64(result.Timestamp))) * 1000000000)
	result.Time = time.Unix(int64(result.Timestamp), nanos)
	return result, nil
}
