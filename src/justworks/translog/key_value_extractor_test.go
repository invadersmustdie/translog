package translog

import (
  "github.com/stretchr/testify/assert"
  "testing"
)

var t_kv_extract_msg = `[08/Sep/2013:11:47:45 +0200] host=proxy01.foo.local direction=c url=/this/is/a/pary?foo=bar&url=%2F http-method=GET http-rc=301 bytes=0 X-Origin= req_start="1.1.1.1 57560 54020166" req_end="54020166 1378633665.745954514 1378633665.746282101 0.000084877 0.000200272 0.000127316" fetch_error="" cookie='foo="{\"ex\":{\"bar\":{\"id\":0,\"gn\":\"A4\",\"x\":1814796400000}}}"; BrowserId=bla; DS=ape=1;' set-cookie='' content_type="" cache_hit=HIT backend="THIS_IS_BACKEND (3.3.3.3:8080)" client_ip= forwarded_for="1.1.1.1" user_agent="Mozilla/4.0 (compatible; MSIE7.0; Windows NT 5.1; Trident/4.0; .NET CLR 1.1.4322; .NET CLR 2.0.50727; .NET CLR 3.0.4506.2152; .NET CLR 3.5.30729; .NET4.0C; .NET4.0E)" cache_hits=0  x-varnish=\"54020166\" vary= incomplete=F`

func Test_KeyValueExtractor_ParseMessage(t *testing.T) {
  msg := t_kv_extract_msg

  e := CreateEvent("test")
  e.SetRawMessage(msg)

  filter := new(KeyValueExtractor)
  fields := filter.ExtractKeyValuePairs(e)

  assert.Equal(t, "proxy01.foo.local", fields["host"])
  assert.Equal(t, "THIS_IS_BACKEND (3.3.3.3:8080)", fields["backend"])
  assert.Equal(t, "0", fields["bytes"])
  assert.Equal(t, `foo="{\"ex\":{\"bar\":{\"id\":0,\"gn\":\"A4\",\"x\":1814796400000}}}"; BrowserId=bla; DS=ape=1;`, fields["cookie"])
  assert.Equal(t, "HIT", fields["cache_hit"])
}

func Test_KeyValueExtractor_MergeExistingFields(t *testing.T) {
  e := CreateEvent("test")
  e.SetRawMessage("fred=wilma")
  e.Fields["barney"] = "betty"

  filter := new(KeyValueExtractor)

  filter.ProcessEvent(e)

  assert.Equal(t, 2, len(e.Fields))
}

func BenchmarkKeyValueExtraction(b *testing.B) {
  for i := 0; i < b.N; i++ {
    e := CreateEvent("test")
    e.SetRawMessage(t_kv_extract_msg)

    filter := new(KeyValueExtractor)
    _ = filter.ExtractKeyValuePairs(e)
  }
}
