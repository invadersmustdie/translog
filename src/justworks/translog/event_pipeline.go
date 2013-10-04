package translog

import (
  "log"
  "runtime"
  "time"
)

type EventPipeline struct {
  ReportInterval  int
  inputProcessor  *InputProcessor
  outputProcessor *OutputProcessor
  filterChain     *FilterChain
  filters         []FilterPlugin
  nMessages       int64
}

func (pipeline *EventPipeline) Init() {
  pipeline.inputProcessor = new(InputProcessor)
  pipeline.filterChain = new(FilterChain)
  pipeline.outputProcessor = new(OutputProcessor)
}

func (pipeline *EventPipeline) AddInput(plugin InputPlugin) {
  pipeline.inputProcessor.RegisterPlugin(plugin)
}

func (pipeline *EventPipeline) AddFilter(plugin FilterPlugin) {
  pipeline.filterChain.RegisterPlugin(plugin)
}

func (pipeline *EventPipeline) AddOutput(plugin ProcessingPlugin) {
  pipeline.outputProcessor.RegisterPlugin(plugin)
}

func (pipeline *EventPipeline) statsRoutine(reportIntervalinSeconds int) {
  var prevCount int64 = 0

  for {
    nMessages := pipeline.nMessages
    newMessage := nMessages - prevCount

    log.Printf(
      "[EventPipeline] %d msg, %d msg/sec, %d total, %d goroutines",
      newMessage,
      (newMessage / int64(reportIntervalinSeconds)),
      nMessages,
      runtime.NumGoroutine())

    prevCount = nMessages
    time.Sleep(time.Duration(reportIntervalinSeconds) * time.Second)
  }
}

func (pipeline *EventPipeline) Start() {
  var in_chan chan *Event = make(chan *Event)

  go pipeline.statsRoutine(pipeline.ReportInterval)

  pipeline.inputProcessor.Start(in_chan)

  for {
    event := <-in_chan

    pipeline.nMessages += 1

    pipeline.filterChain.ProcessEvent(event)

    if event.KeepEvent {
      pipeline.outputProcessor.ProcessEvent(event)
    }
  }
}
