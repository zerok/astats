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
	line := `{"level":"info","ts":1589436543.3458548,"logger":"http.log.access.log0","msg":"handled request","request":{"method":"GET","uri":"/","proto":"HTTP/2.0","remote_addr":"127.0.0.1:27615","host":"zerokspot.com","headers":{"User-Agent":["Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:77.0) Gecko/20100101 Firefox/77.0"],"Accept-Encoding":["gzip, deflate, br"],"Upgrade-Insecure-Requests":["1"],"If-Modified-Since":["Thu, 14 May 2020 06:06:28 GMT"],"Accept":["text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"],"Accept-Language":["en-US,en;q=0.5"],"If-None-Match":["\"qab4ys83i\""],"Cache-Control":["max-age=0"],"Te":["trailers"]},"tls":{"resumed":false,"version":772,"ciphersuite":4865,"proto":"h2","proto_mutual":true,"server_name":"zerokspot.com"}},"common_log":"80.110.127.9 - - [14/May/2020:06:09:03 +0000] \"GET / HTTP/2.0\" 304 0","duration":0.000221397,"size":0,"status":304,"resp_headers":{"Server":["Caddy"],"Etag":["\"qab4ys83i\""]}}
`
	lf := &AccessLogFile{}
	require.NoError(t, lf.InitFromReader(bytes.NewBufferString(line)))
	readLine, err := lf.NextLine(ctx)
	require.NoError(t, err)
	require.Equal(t, "/", readLine.Request.URI)
	require.NoError(t, err)
	require.Equal(t, time.Date(2020, time.May, 14, 6, 9, 3, 345854759, time.UTC), readLine.Time.In(time.UTC))
	_, err = lf.NextLine(ctx)
	require.Error(t, err)
}
