package translog

import (
  "github.com/stretchr/testify/assert"
  "testing"
)

func TestCreateWithSource(t *testing.T) {
  event := CreateEvent("file://foobar")

  assert.Equal(t, "file://foobar", event.Source)
}
