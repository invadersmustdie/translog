package translog

type InputPlugin interface {
  Configure(config map[string]string)
  Start(c chan *Event)
}
