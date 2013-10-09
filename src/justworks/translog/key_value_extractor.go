package translog

import (
  "log"
  "regexp"
)

type KeyValueExtractor struct {
  debug bool
}

func (filter *KeyValueExtractor) Configure(config map[string]string) {
  if config["debug"] == "true" {
    filter.debug = true
  }
}

func (filter *KeyValueExtractor) ProcessEvent(e *Event) {
  e.Fields = MergeMap(filter.ExtractKeyValuePairs(e), e.Fields)
}

func (filter *KeyValueExtractor) ExtractKeyValuePairs(e *Event) map[string]string {
  msg := e.RawMessage

  date_re, _ := regexp.Compile(`\[([^\]]+)\]`)
  field_re, _ := regexp.Compile(`([A-Za-z_-]+)=(\"([^"]*)\"|\'([^']*)\'|\S+)`)

  fields := make(map[string]string)

  date_match := date_re.FindAllStringSubmatch(msg, -1)
  field_match := field_re.FindAllStringSubmatch(msg, -1)

  for group_idx := 0; group_idx < len(field_match); group_idx++ {
    group := field_match[group_idx]

    value := ""
    for i := 0; i < len(group); i++ {
      if len(group[i]) > 0 {
        value = group[i]
      }
    }

    if filter.debug {
      log.Printf("[%T] extracted [%s]='%s'", filter, group[1], value)
    }

    fields[group[1]] = value
  }

  if len(date_match) > 0 && len(date_match[0]) > 0 {
    fields["date"] = date_match[0][1]
  }

  return fields
}
