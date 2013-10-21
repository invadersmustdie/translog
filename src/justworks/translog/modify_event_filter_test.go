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

func Test_ModifyEventFilter_substitute_simple(t *testing.T) {
  e := CreateEvent("test")
  e.SetRawMessage("[xasd] flsl fii")

  filter := new(ModifyEventFilter)
  filter.Configure(map[string]string{
    "msg.replace.pattern":    "fii",
    "msg.replace.substitute": "",
    "debug":                  "true",
  })

  new_event := filter.Modify(e)

  assert.Equal(t, "[xasd] flsl ", new_event.RawMessage)
}

func Test_ModifyEventFilter_substitute_simple_no_substitute(t *testing.T) {
  e := CreateEvent("test")
  e.SetRawMessage("[xasd] flsl fii")

  filter := new(ModifyEventFilter)
  filter.Configure(map[string]string{
    "msg.replace.pattern": "fii",
    "debug":               "true",
  })

  new_event := filter.Modify(e)

  assert.Equal(t, "[xasd] flsl ", new_event.RawMessage)
}

func Test_ModifyEventFilter_substitute_simple_with_replacement(t *testing.T) {
  e := CreateEvent("test")
  e.SetRawMessage("[xasd] flsl fii")

  filter := new(ModifyEventFilter)
  filter.Configure(map[string]string{
    "msg.replace.pattern":    "fii",
    "msg.replace.substitute": "bar",
    "debug":                  "true",
  })

  new_event := filter.Modify(e)

  assert.Equal(t, "[xasd] flsl bar", new_event.RawMessage)
}

func Test_ModifyEventFilter_substitute_simple_with_backref(t *testing.T) {
  e := CreateEvent("test")
  e.SetRawMessage("[xasd] foo='my_value002002;sklsllsls' bla='baz'")

  filter := new(ModifyEventFilter)
  filter.Configure(map[string]string{
    "msg.replace.pattern":    "foo='([a-zA-Z0-9_]+)[^']*'",
    "msg.replace.substitute": "foo='$1'",
    "debug":                  "true",
  })

  new_event := filter.Modify(e)

  assert.Equal(t, "[xasd] foo='my_value002002' bla='baz'", new_event.RawMessage)
}
