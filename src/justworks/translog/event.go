package translog

import (
  "fmt"
  "log"
  "os"
  "time"
)

type Event struct {
  InitTime   time.Time
  Time       time.Time
  RawMessage string
  Host       string
  Source     string
  KeepEvent  bool
  Fields     map[string]string
}

func CreateEvent(source string) *Event {
  event := new(Event)

  event.InitTime = time.Now()
  event.Time = time.Now()
  event.KeepEvent = true
  event.Fields = make(map[string]string)

  hostname, err := os.Hostname()

  if err != nil {
    if debug {
      log.Printf("Failed obtaining hostname while creating new event, using localhost as fallback")
    }
    hostname = "localhost"
  }

  event.Host = hostname
  event.Source = source

  return event
}

func (e *Event) SetRawMessage(msg string) *Event {
  e.RawMessage = msg

  return e
}

func (e *Event) SetTime(time time.Time) *Event {
  e.Time = time

  return e
}

func (e *Event) PrettyPrint() string {
  pp_string := "<#Event>" +
    fmt.Sprintf("  Host: %s\n", e.Host) +
    fmt.Sprintf("  Source: %s\n", e.Source) +
    fmt.Sprintf("  InitTime: %s\n", e.InitTime) +
    fmt.Sprintf("  Time: %s\n", e.Time) +
    fmt.Sprintf("  Raw:\n") +
    fmt.Sprintf("%s\n", e.RawMessage) +
    fmt.Sprintf("  Fields:\n")

  for k, v := range e.Fields {
    pp_string += fmt.Sprintf("     [%s] %s\n", k, v)
  }

  return pp_string
}
