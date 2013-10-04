package translog

import (
  "github.com/stretchr/testify/assert"
  "testing"
)

func Test_ModifyEventFilter_removeFields(t *testing.T) {
  e := CreateEvent("test")
  e.Fields["bar"] = "x1"
  e.Fields["foo"] = "x2"
  e.Fields["baz"] = "x3"

  filter := new(ModifyEventFilter)
  filter.Configure(map[string]string{
    "field.remove.list": "bar,foo",
  })

  new_event := filter.Modify(e)

  assert.NotNil(t, new_event.Fields["baz"])
  assert.Equal(t, 1, len(new_event.Fields))
}

func Test_ModifyEventFilter_removeFields_withSpace(t *testing.T) {
  e := CreateEvent("test")
  e.Fields["bar"] = "x1"
  e.Fields["foo"] = "x2"
  e.Fields["baz"] = "x3"

  filter := new(ModifyEventFilter)
  filter.Configure(map[string]string{
    "field.remove.list": "bar, foo",
  })

  new_event := filter.Modify(e)

  assert.NotNil(t, new_event.Fields["baz"])
  assert.Equal(t, 1, len(new_event.Fields))
}

func Test_ModifyEventFilter_doNothing(t *testing.T) {
  e := CreateEvent("test")

  filter := new(ModifyEventFilter)
  filter.Configure(map[string]string{
    "field.remove.list": "bar,foo",
  })

  new_event := filter.Modify(e)

  assert.Equal(t, 0, len(new_event.Fields))
}

func Test_ModifyEventFilter_removeFields_byPattern(t *testing.T) {
  e := CreateEvent("test")
  e.Fields["bar"] = "x1"
  e.Fields["foo"] = "x2"
  e.Fields["baz"] = "x3"

  filter := new(ModifyEventFilter)
  filter.Configure(map[string]string{
    "field.remove.match": "^b",
  })

  new_event := filter.Modify(e)

  assert.NotNil(t, new_event.Fields["foo"])
  assert.Equal(t, 1, len(new_event.Fields))
}
