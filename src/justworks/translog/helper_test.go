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

func Test_Helper_MergeMap_Simple(t *testing.T) {
  m1 := map[string]string{
    "fred": "wilma",
  }

  m2 := map[string]string{
    "barney": "betty",
  }

  result := MergeMap(m1, m2)

  assert.Equal(t, 2, len(result))
  assert.Equal(t, "wilma", result["fred"])
  assert.Equal(t, "betty", result["barney"])
}

func Test_Helper_MergeMap_ConflictingKey_in_Source(t *testing.T) {
  m1 := map[string]string{
    "fred":   "wilma",
    "barney": "betty",
  }

  m2 := map[string]string{
    "barney": "not",
  }

  result := MergeMap(m1, m2)

  assert.Equal(t, 2, len(result))
  assert.Equal(t, "wilma", result["fred"])
  assert.Equal(t, "betty", result["barney"])
}

func Test_Helper_MergeMap_ConflictingKey_in_Target(t *testing.T) {
  m1 := map[string]string{
    "fred": "wilma",
  }

  m2 := map[string]string{
    "barney": "betty",
    "fred":   "not",
  }

  result := MergeMap(m1, m2)

  assert.Equal(t, 2, len(result))
  assert.Equal(t, "wilma", result["fred"])
  assert.Equal(t, "betty", result["barney"])
}
