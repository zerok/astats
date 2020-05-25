package accesslog

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseLine(t *testing.T) {
	ctx := context.Background()
	line := `13.66.139.0 - - [06/May/2020:17:57:56 +0000] "GET /index.xml HTTP/1.1" 200 21604 "-" "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)" "-"`
	lf := &AccessLogFile{}
	require.NoError(t, lf.InitFromReader(bytes.NewBufferString(line)))
	readLine, err := lf.NextLine(ctx)
	require.NoError(t, err)
	require.Equal(t, "/index.xml", readLine.Path)
	loc := time.FixedZone("+0000", 0)
	require.Equal(t, time.Date(2020, time.May, 6, 17, 57, 56, 0, loc), readLine.Time)
	_, err = lf.NextLine(ctx)
	require.Error(t, err)
}
