package translog

import (
  "log"
  "regexp"
  "strings"
)

type DropEventFilter struct {
  fieldMatches       map[string]regexp.Regexp
  msgMatch           regexp.Regexp
  msgMatchIsNegative bool
  debug              bool
}

func (filter *DropEventFilter) Configure(config map[string]string) {
  log.Printf("[%T] config %v", filter, config)
  filter.fieldMatches = make(map[string]regexp.Regexp)

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
        log.Printf("[%T] added field filter [%s]='%s'", filter, fieldName, v)
      }

      filter.fieldMatches[fieldName] = *filterRegex
    }
  }

  for k, v := range filter.fieldMatches {
    log.Printf("%s -> %s", k, v)
  }

  if len(config["msg.match"]) > 0 {
    pattern := config["msg.match"]

    if strings.HasPrefix(pattern, "!") {
      log.Printf("[%T] found '!' prefix, changing to negative match", filter)
      filter.msgMatchIsNegative = true
      pattern = strings.Replace(pattern, "!", "", 1)
    }

    re, err := regexp.Compile(pattern)

    if err != nil {
      log.Printf("[%T] failed compiling regex for msg.match (regex=%s) (%s)", filter, config["msg.match"], err)
    }

    filter.msgMatch = *re
  }
}

func (filter *DropEventFilter) ProcessEvent(e *Event) {
  for fieldName, fieldPattern := range filter.fieldMatches {
    fieldValue := e.Fields[fieldName]

    if len(fieldValue) == 0 {
      // set fieldValue to empty string if field is not set in event
      fieldValue = ""
    }

    if fieldPattern.MatchString(fieldValue) {
      log.Printf("[%T] dropping event because value %s of field %s matches %s", filter, fieldValue, fieldName, fieldPattern)

      e.KeepEvent = false
      return
    }
  }

  if len(filter.msgMatch.String()) > 0 {
    hasMatch := filter.msgMatch.MatchString(e.RawMessage)

    if (hasMatch && !filter.msgMatchIsNegative) || (!hasMatch && filter.msgMatchIsNegative) {
      log.Printf("[%T] dropping event because raw message matches %s", filter, filter.msgMatch.String())
      e.KeepEvent = false
      return
    }
  }
}
