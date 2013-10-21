package translog

import (
  "fmt"
  "log"
  "strings"
)

type StatsdPlugin struct {
  socketWriter NetworkSocketWriter
  placeholders []string
  config       map[string]string
  debug        bool
  host         string
}

func (plugin *StatsdPlugin) Configure(config map[string]string) {
  plugin.config = config

  for k, v := range config {
    if strings.HasPrefix(k, "field.") && strings.HasSuffix(k, ".raw") {
      plugin.placeholders = append(plugin.placeholders, v)
    }
  }

  if config["debug"] == "true" {
    plugin.debug = true
  }

  w := CreateNetworkSocketWriter(plugin, plugin.config)
  plugin.socketWriter = w
}

func (plugin *StatsdPlugin) ExtractMetrics(e *Event) []string {
  return FieldsWithReplacedPlaceholders(e, plugin.placeholders, fmt.Sprintf("%V", plugin), plugin.debug)
}

func (plugin *StatsdPlugin) ProcessEvent(e *Event) {
  metrics := plugin.ExtractMetrics(e)

  for _, metric := range metrics {
    m := fmt.Sprintf("%s", metric)

    if plugin.debug {
      log.Printf("[%T] sending %s", plugin, metric)
    }

    plugin.socketWriter.WriteString(m)
  }
}
