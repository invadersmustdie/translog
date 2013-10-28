package translog

import (
  "log"
  "regexp"
  "strings"
)

type ModifyEventFilter struct {
  removeFields       []string
  removeFieldPattern regexp.Regexp
  replacePattern     regexp.Regexp
  replaceSubstitute  string
  additionalFields   map[string]string
  debug              bool
}

func (filter *ModifyEventFilter) Configure(config map[string]string) {
  log.Printf("[%T] config %v", filter, config)

  filter.additionalFields = make(map[string]string)

  if config["debug"] == "true" {
    filter.debug = true
  }

  /* field.remove.list */
  if len(config["field.remove.list"]) > 0 {
    fields := strings.Split(config["field.remove.list"], ",")

    for _, f := range fields {
      trimedString := strings.TrimSpace(f)
      filter.removeFields = append(filter.removeFields, trimedString)
    }
  }

  /* field.add */
  for k, v := range config {
    if strings.HasPrefix(k, "field.add.") {
      field := k[(strings.LastIndex(k, ".")+1):]

      if len(field) > 0 {
        filter.additionalFields[field] = v
      }
    }
  }

  /* field.remove.match */
  removeFieldPattern := config["field.remove.match"]

  if len(removeFieldPattern) > 0 {
    pattern, err := regexp.Compile(removeFieldPattern)

    if err != nil {
      log.Printf("[%T] failed compiling pattern %s (%s)", filter, removeFieldPattern, err)
    }

    filter.removeFieldPattern = *pattern
  }

  /* msg.replace.pattern */
  if len(config["msg.replace.pattern"]) > 0 {
    pattern, err := regexp.Compile(config["msg.replace.pattern"])

    if err != nil {
      log.Printf("[%T] failed compiling pattern %s (%s)", filter, config["msg.replace.pattern"], err)
    }

    filter.replacePattern = *pattern

    if len(config["msg.replace.pattern"]) > 0 {
      filter.replaceSubstitute = config["msg.replace.substitute"]
    }
  }
}

func (filter *ModifyEventFilter) ProcessEvent(e *Event) {
  filter.Modify(e)
}

func (filter *ModifyEventFilter) Modify(e *Event) Event {
  for _, fieldToRemove := range filter.removeFields {
    if len(e.Fields[fieldToRemove]) > 0 {
      if filter.debug {
        log.Printf("[%T] delete %s", filter, fieldToRemove)
      }
      delete(e.Fields, fieldToRemove)
    }
  }

  if len(filter.removeFieldPattern.String()) > 0 {
    for fieldName, _ := range e.Fields {
      if filter.removeFieldPattern.MatchString(fieldName) {
        if filter.debug {
          log.Printf("[%T] removing field %s", filter, fieldName)
        }
        delete(e.Fields, fieldName)
      }
    }
  }

  if len(filter.replacePattern.String()) > 0 {
    if filter.replacePattern.MatchString(e.RawMessage) {
      result := filter.replacePattern.ReplaceAllString(e.RawMessage, filter.replaceSubstitute)

      if filter.debug {
        log.Printf("[%T] message replacement pattern='%s' substitute='%s' before='%s' after='%s'",
          filter,
          filter.replacePattern.String(),
          filter.replaceSubstitute,
          e.RawMessage,
          result)
      }

      e.SetRawMessage(result)
    }
  }

  for k,v := range filter.additionalFields {
    if len(e.Fields[k]) == 0 {
      if filter.debug {
        log.Printf("[%T] adding field='%s' value='%s'", k, v)
      }

      e.Fields[k] = v
    }
  }

  return *e
}
