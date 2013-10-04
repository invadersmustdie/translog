package translog

import (
  "fmt"
)

type StdoutWriterPlugin struct {
  config map[string]string
}

func (plugin *StdoutWriterPlugin) Configure(config map[string]string) {
  plugin.config = config
}

func (plugin *StdoutWriterPlugin) ProcessEvent(event *Event) {
  fmt.Print(event.PrettyPrint())
}
