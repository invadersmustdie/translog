package translog

import (
  "log"
  "regexp"
  "strings"
)

type ModifyEventFilter struct {
  removeFields       []string
  removeFieldPattern regexp.Regexp
  debug              bool
}

func (filter *ModifyEventFilter) Configure(config map[string]string) {
  log.Printf("[%T] config %v", filter, config)

  if len(config["field.remove.list"]) > 0 {
    fields := strings.Split(config["field.remove.list"], ",")

    for _, f := range fields {
      trimedString := strings.TrimSpace(f)
      filter.removeFields = append(filter.removeFields, trimedString)
    }
  }

  removeFieldPattern := config["field.remove.match"]

  if len(removeFieldPattern) > 0 {
    pattern, err := regexp.Compile(removeFieldPattern)

    if err != nil {
      log.Printf("[%T] failed compiling pattern %s (%s)", filter, removeFieldPattern, err)
    }

    filter.removeFieldPattern = *pattern
  }

  if config["debug"] == "true" {
    filter.debug = true
  }
}

func (filter *ModifyEventFilter) ProcessEvent(e *Event) {
  filter.Modify(e)
}

func (filter *ModifyEventFilter) Modify(e *Event) Event {
  for _, fieldToRemove := range filter.removeFields {
    if len(e.Fields[fieldToRemove]) > 0 {
      if filter.debug {
        log.Printf("delete %s", fieldToRemove)
      }
      delete(e.Fields, fieldToRemove)
    }
  }

  if len(filter.removeFieldPattern.String()) > 0 {
    for fieldName, _ := range e.Fields {
      if filter.removeFieldPattern.MatchString(fieldName) {
        if filter.debug {
          log.Printf("[%T] removing field %s", fieldName)
        }
        delete(e.Fields, fieldName)
      }
    }
  }

  return *e
}
