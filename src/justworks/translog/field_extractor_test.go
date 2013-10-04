package translog

import (
  "github.com/stretchr/testify/assert"
  "testing"
)

func TestExtractField(t *testing.T) {
  msg := `[08/Sep/2013:11:47:45 +0200] host=proxy01 direction=c url=/foo/bar?bla=xx http-method=GET http-rc=301 bytes=0 X-Origin= req_start="1.1.1.1 57560 54020166" req_end="54020166 1378633665.745954514 1378633665.746282101 0.000084877 0.000200272 0.000127316" fetch_error="" cookie='visitorId=e835c6a97622cdc26b9095caebd2df14b5a863fce00df466ed17353dded48df16cf57733f860e897386103889298387b; uniqueUserId=599523f2-9962-4ed0-9994-0ad1b2973e52; BrowserXId=ddf34e8e200b933bfdd5ae79988beee775baf9755c20dec934edb187628c1fb270d14e5e1a876f3e12138d57b978a2bc; foo=ape=1;' set-cookie='' content_type="" cache_hit= backend="" client_ip= forwarded_for="2.2.2.2" user_agent="Mozilla/4.0 (compatible; MSIE7.0; Windows NT 5.1; Trident/4.0; .NET CLR 1.1.4322; .NET CLR 2.0.50727; .NET CLR 3.0.4506.2152; .NET CLR 3.5.30729; .NET4.0C; .NET4.0E)" cache_hits=0  x-varnish=\"54020166\" vary= incomplete=F`

  e := CreateEvent("test")
  e.SetRawMessage(msg)

  filter := new(FieldExtractor)
  filter.Configure(map[string]string{
    "field.visitorId": "visitorId=([a-zA-Z0-9]+);",
  })
  extracted_fields := filter.ExtractFieldByPattern(e)

  assert.Equal(t, "e835c6a97622cdc26b9095caebd2df14b5a863fce00df466ed17353dded48df16cf57733f860e897386103889298387b", extracted_fields["visitorId"])
}

func TestExtractFieldNoMatch(t *testing.T) {
  msg := `foo`

  e := CreateEvent("test")
  e.SetRawMessage(msg)

  filter := new(FieldExtractor)
  filter.Configure(map[string]string{
    "field.visitorId": "visitorId=([a-zA-Z0-9]+);",
  })
  extracted_fields := filter.ExtractFieldByPattern(e)

  assert.Equal(t, "", extracted_fields["visitorId"])
}
