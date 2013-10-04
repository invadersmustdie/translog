package translog

type ProcessingPlugin interface {
  ProcessEvent(event *Event)
  Configure(config map[string]string)
}
