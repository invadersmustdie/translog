package translog

import (
  "github.com/stretchr/testify/assert"
  "testing"
)

func Test_StatsdPlugin_ExtractCacheHit(t *testing.T) {
  filter := new(StatsdPlugin)
  filter.Configure(map[string]string{
    "host":        "statsd.foo.local",
    "port":        "123",
    "proto":       "udp",
    "field.0.raw": "foo.bar.cache_hit.%{cache_hit}:1|c",
  })

  event := CreateEvent("test")
  event.SetRawMessage("")
  event.Fields["cache_hit"] = "MISS"

  metrics := filter.ExtractMetrics(event)

  assert.Equal(t, 1, len(metrics))
  assert.Equal(t, "foo.bar.cache_hit.MISS:1|c", metrics[0])
}

func Test_StatsdPlugin_ExtractNothing(t *testing.T) {
  filter := new(StatsdPlugin)
  filter.Configure(map[string]string{
    "host":        "statsd.foo.local",
    "port":        "123",
    "proto":       "udp",
    "field.0.raw": "foo.bar.cache_hit.%{cache_hit}:1|c",
  })

  event := CreateEvent("test")
  event.SetRawMessage("")

  metrics := filter.ExtractMetrics(event)

  assert.Equal(t, 0, len(metrics))
}

func Test_StatsdPlugin_ExtractTwoFieldIntoOneMetric(t *testing.T) {
  filter := new(StatsdPlugin)
  filter.Configure(map[string]string{
    "debug":       "true",
    "host":        "statsd.foo.local",
    "port":        "123",
    "proto":       "udp",
    "field.0.raw": "foo.bar.%{type}.cache_hit.%{cache_hit}:1|c",
  })

  event := CreateEvent("test")
  event.SetRawMessage("")
  event.Fields["cache_hit"] = "MISS"
  event.Fields["type"] = "DIRECT"

  metrics := filter.ExtractMetrics(event)

  assert.Equal(t, 1, len(metrics))
  assert.Equal(t, "foo.bar.DIRECT.cache_hit.MISS:1|c", metrics[0])
}
