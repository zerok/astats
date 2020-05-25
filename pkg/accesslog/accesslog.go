package accesslog

import (
	"bufio"
	"context"
	"io"
	"strconv"
	"time"

	"github.com/vjeantet/grok"
)

type LineFormat struct {
	Format string
}

type AccessLogFile struct {
	Format          LineFormat
	TimestampFormat string
	Path            string
	r               *bufio.Reader
	g               *grok.Grok
}

type LogEntry struct {
	Time       time.Time
	StatusCode int
	SourceIP   string
	Path       string
	UserAgent  string
	Referrer   string
}

func (lf *AccessLogFile) InitFromReader(r io.Reader) error {
	lf.r = bufio.NewReader(r)
	lf.g, _ = grok.New()
	lf.g.AddPattern("TIME_LOCAL", "%{MONTHDAY}/%{MONTH}/%{YEAR}:%{HOUR}:%{MINUTE}:%{SECOND} %{ISO8601_TIMEZONE}")
	return nil
}

func (lf *AccessLogFile) NextLine(ctx context.Context) (*LogEntry, error) {
	line, err := lf.r.ReadString('\n')
	if err != nil {
		if line == "" {
			return nil, err
		}
	}
	result := LogEntry{}
	values, err := lf.g.Parse("%{IP:ip} - %{DATA:remote_user} \\[%{TIME_LOCAL:timestamp}\\] \"%{DATA:verb} %{DATA:path} %{DATA:http_version}\" %{DATA:response_code} %{DATA:response_size} \"%{DATA:referrer}\" \"%{DATA:user_agent}\"", line)
	if err != nil {
		return nil, err
	}
	result.Path = values["path"]
	result.SourceIP = values["ip"]
	result.UserAgent = values["user_agent"]
	result.Referrer = values["referrer"]
	if values["response_code"] != "" {
		result.StatusCode, err = strconv.Atoi(values["response_code"])
		if err != nil {
			return nil, err
		}
	}
	if values["timestamp"] != "" {
		ts, err := time.Parse("02/Jan/2006:15:04:05 MST", values["timestamp"])
		if err != nil {
			return nil, err
		}
		result.Time = ts
	}
	return &result, nil
}
