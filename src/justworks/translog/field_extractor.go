package translog

import (
  "log"
  "regexp"
  "strings"
)

type FieldExtractor struct {
  fieldFilter map[string]regexp.Regexp
  debug       bool
}

func (filter *FieldExtractor) Configure(config map[string]string) {
  log.Printf("[%T] config %v", filter, config)
  filter.fieldFilter = make(map[string]regexp.Regexp)

  if config["debug"] == "true" {
    filter.debug = true
  }

  for k, v := range config {
    if strings.HasPrefix(k, "field.") {
      fieldName := strings.Replace(k, "field.", "", 1)

      filterRegex, err := regexp.Compile(v)

      if err != nil {
        log.Printf("[%T] failed compiling regex for %s (regex=%s) (%s)", filter, k, v, err)
        continue
      }

      if filter.debug {
        log.Printf("[%T] adding pattern %s", filter, v)
      }

      filter.fieldFilter[fieldName] = *filterRegex
    }
  }
}

func (filter *FieldExtractor) ProcessEvent(e *Event) {
  new_fields := filter.ExtractFieldByPattern(e)

  for k, v := range new_fields {
    e.Fields[k] = v
  }
}

func (filter *FieldExtractor) ExtractFieldByPattern(e *Event) map[string]string {
  fields := make(map[string]string)

  for fieldName, fieldPattern := range filter.fieldFilter {
    match := fieldPattern.FindAllStringSubmatch(e.RawMessage, -1)

    if len(match) > 0 && len(match[0]) > 1 && len(match[0][0]) > 0 {
      if filter.debug {
        log.Printf("[%T] field[%s] matched value='%v'", filter, fieldName, match[0][1])
      }

      fields[fieldName] = match[0][1]
    }
  }

  return fields
}
