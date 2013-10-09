package translog

import (
  "log"
  "regexp"
  "strings"
)

func FieldsWithReplacedPlaceholders(e *Event, placeholders []string, caller string, debug bool) []string {
  var metrics []string

  field_placeholder_re, _ := regexp.Compile(`%{[^\s\}]+}`)

  for _, r := range placeholders {
    metric := ""
    matches := field_placeholder_re.FindAllStringSubmatch(r, -1)
    fields := []string{}

    for _, m := range matches {
      fields = append(fields, m[0])
    }

    for idx, m := range fields {
      fieldName := strings.Trim(m, "%{}")

      if debug {
        log.Printf("[%s] %s -> found matching field %d %v %v", caller, r, idx, m, fieldName)
      }

      if len(e.Fields[fieldName]) > 0 {
        if debug {
          log.Printf("[%s] %s -> replacing field field %d %v %v", caller, r, idx, m, fieldName)
        }

        val := r
        if len(metric) > 0 {
          val = metric
        }

        metric = strings.Replace(val, m, e.Fields[fieldName], -1)
      }
    }

    if len(metric) > 0 {
      if debug {
        log.Printf("[%s] adding metric %s", caller, metric)
      }

      metrics = append(metrics, metric)
    }
  }

  return metrics
}
