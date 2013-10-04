package translog

import (
  "log"
)

type InputProcessor struct {
  Plugins []InputPlugin
}

func (inputProcessor *InputProcessor) RegisterPlugin(plugin InputPlugin) {
  log.Printf("[inputProcessor] register %T\n", plugin)
  inputProcessor.Plugins = append(inputProcessor.Plugins, plugin)
}

func (inputProcessor *InputProcessor) Start(c chan *Event) {
  for _, plugin := range inputProcessor.Plugins {
    debug.Printf("Starting %T\n", plugin)
    go plugin.Start(c)
  }
}
