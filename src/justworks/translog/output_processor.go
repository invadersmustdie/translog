package translog

import (
  "log"
)

type OutputProcessor struct {
  Plugins []ProcessingPlugin
}

func (outputProcessor *OutputProcessor) RegisterPlugin(plugin ProcessingPlugin) {
  log.Printf("[outputProcessor] register %T\n", plugin)
  outputProcessor.Plugins = append(outputProcessor.Plugins, plugin)
}

func (outputProcessor *OutputProcessor) ProcessEvent(event *Event) {
  for _, plugin := range outputProcessor.Plugins {
    debug.Printf("[outputProcessor] Dispatching event to %T\n", plugin)
    go plugin.ProcessEvent(event)
  }
}
