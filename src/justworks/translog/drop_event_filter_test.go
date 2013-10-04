package translog

import (
  "github.com/stretchr/testify/assert"
  "testing"
)

func Test_DropByMatchField(t *testing.T) {
  e := CreateEvent("test")
  e.Fields["foo"] = "bar"

  filter := new(DropEventFilter)
  filter.Configure(map[string]string{
    "field.foo": "^bar",
  })

  filter.ProcessEvent(e)

  assert.False(t, e.KeepEvent)
}

func Test_DoNotDropByMatchField(t *testing.T) {
  e := CreateEvent("test")
  e.Fields["foo"] = "bar"

  filter := new(DropEventFilter)
  filter.Configure(map[string]string{
    "field.foo": "lalal",
  })

  filter.ProcessEvent(e)

  assert.True(t, e.KeepEvent)
}

func Test_DropByMatchContent(t *testing.T) {
  e := CreateEvent("test")
  e.SetRawMessage("foobar")

  filter := new(DropEventFilter)
  filter.Configure(map[string]string{
    "msg.match": "foo",
  })

  filter.ProcessEvent(e)

  assert.False(t, e.KeepEvent)
}

func Test_DoNotDropByMatchContent(t *testing.T) {
  e := CreateEvent("test")
  e.SetRawMessage("foobar")

  filter := new(DropEventFilter)
  filter.Configure(map[string]string{
    "msg.match": "xx",
  })

  filter.ProcessEvent(e)

  assert.True(t, e.KeepEvent)
}
