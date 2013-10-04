package translog

type FilterPlugin interface {
  ProcessEvent(e *Event)
  Configure(config map[string]string)
}
