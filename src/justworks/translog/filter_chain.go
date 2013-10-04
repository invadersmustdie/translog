package translog

import (
  "log"
)

type FilterChain struct {
  Plugins []FilterPlugin
}

func (filterChain *FilterChain) RegisterPlugin(plugin FilterPlugin) {
  log.Printf("[filterChain] register %T\n", plugin)
  filterChain.Plugins = append(filterChain.Plugins, plugin)
}

func (filterChain *FilterChain) ProcessEvent(event *Event) {
  for _, plugin := range filterChain.Plugins {
    debug.Printf("[filterChain] Dispatching event to %T\n", plugin)

    if event.KeepEvent {
      plugin.ProcessEvent(event)
    } else {
      debug.Printf("[%T] Dropped Event because KeepEvent=false")
      return
    }
  }
}
