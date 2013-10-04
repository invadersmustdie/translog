package translog

import (
  "fmt"
  "log"
  "regexp"
  "strings"
)

type StatsdPlugin struct {
  socketWriter NetworkSocketWriter
  rawRequests  []string
  config       map[string]string
  debug        bool
  host         string
}

func (plugin *StatsdPlugin) Configure(config map[string]string) {
  plugin.config = config

  for k, v := range config {
    if strings.HasPrefix(k, "field.") && strings.HasSuffix(k, ".raw") {
      plugin.rawRequests = append(plugin.rawRequests, v)
    }
  }

  if config["debug"] == "true" {
    plugin.debug = true
  }

  w := CreateNetworkSocketWriter(plugin, plugin.config)
  plugin.socketWriter = w
}

func (plugin *StatsdPlugin) ExtractMetrics(e *Event) []string {
  var metrics []string

  field_placeholder_re, _ := regexp.Compile(`%{[^\s\}]+}`)

  for _, r := range plugin.rawRequests {
    metric := ""
    matches := field_placeholder_re.FindAllStringSubmatch(r, -1)
    fields := []string{}

    for _, m := range matches {
      fields = append(fields, m[0])
    }

    for idx, m := range fields {
      fieldName := strings.Trim(m, "%{}")

      if plugin.debug {
        log.Printf("[%T] %s -> found matching field %d %v %v", plugin, r, idx, m, fieldName)
      }

      if len(e.Fields[fieldName]) > 0 {
        if plugin.debug {
          log.Printf("[%T] %s -> replacing field field %d %v %v", plugin, r, idx, m, fieldName)
        }

        val := r
        if len(metric) > 0 {
          val = metric
        }

        metric = strings.Replace(val, m, e.Fields[fieldName], -1)
      }
    }

    if len(metric) > 0 {
      if plugin.debug {
        log.Printf("[%T] adding metric %s", plugin, metric)
      }

      metrics = append(metrics, metric)
    }
  }

  return metrics
}

func (plugin *StatsdPlugin) ProcessEvent(e *Event) {
  metrics := plugin.ExtractMetrics(e)

  for _, metric := range metrics {
    m := fmt.Sprintf("%s\n", metric)

    if plugin.debug {
      log.Printf("[%T] sending %s", plugin, metric)
    }

    plugin.socketWriter.WriteString(m)
  }
}
