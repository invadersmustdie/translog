package translog

import (
  "github.com/stretchr/testify/assert"
  "testing"
)

func Test_FieldsWithReplacedPlaceholders_withOneMatch(t *testing.T) {
  e := CreateEvent("foo:///")
  e.Fields["fred"] = "wilma"

  placeholders := []string{"lala.%{fred}:xx"}

  result_list := FieldsWithReplacedPlaceholders(e, placeholders, "none", false)

  assert.Equal(t, 1, len(result_list))
  assert.Equal(t, "lala.wilma:xx", result_list[0])
}

func Test_FieldsWithReplacedPlaceholders_withTwoMatches(t *testing.T) {
  e := CreateEvent("foo:///")
  e.Fields["fred"] = "wilma"
  e.Fields["barney"] = "betty"

  placeholders := []string{"lala.%{fred}:xx", "foooo.%{barney}"}

  result_list := FieldsWithReplacedPlaceholders(e, placeholders, "none", false)

  assert.Equal(t, 2, len(result_list))
  assert.Equal(t, "lala.wilma:xx", result_list[0])
  assert.Equal(t, "foooo.betty", result_list[1])
}

func Test_FieldsWithReplacedPlaceholders_withNoMatch(t *testing.T) {
  e := CreateEvent("foo:///")
  e.Fields["fred"] = "wilma"

  placeholders := []string{"lala.%{fredx}:xx"}

  result_list := FieldsWithReplacedPlaceholders(e, placeholders, "none", false)

  assert.Equal(t, 0, len(result_list))
}

func Test_FieldsWithReplacedPlaceholders_withNoPlaceholder(t *testing.T) {
  e := CreateEvent("foo:///")

  placeholders := []string{"lala.xx"}

  result_list := FieldsWithReplacedPlaceholders(e, placeholders, "none", false)

  assert.Equal(t, 0, len(result_list))
}
