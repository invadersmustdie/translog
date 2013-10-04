package translog

import (
  "github.com/stretchr/testify/assert"
  "testing"
)

func Test_GelfWriterPlugin_CreateMessage(t *testing.T) {
  e := CreateEvent("test")
  e.Fields["bar"] = "x1"
  e.Fields["foo"] = "x2"
  e.Fields["baz"] = "x3"

  plugin := new(GelfWriterPlugin)
  msg := plugin.CreateGelfMessage(e)

  assert.NotNil(t, msg.GelfStr)
  assert.Equal(t, e, msg.Event)
}
