package translog

import (
  "fmt"
  "log"
  "os"
  "syscall"
)

type NamedPipeWriterPlugin struct {
  config map[string]string
  debug  bool
}

func (plugin *NamedPipeWriterPlugin) Configure(config map[string]string) {
  plugin.config = config

  if config["debug"] == "true" {
    plugin.debug = true
  }

  if len(config["filename"]) < 1 {
    log.Fatalf("[%T] ERROR: missing configuration option 'filename'", plugin)
  }
}

func (plugin *NamedPipeWriterPlugin) ProcessEvent(event *Event) {
  config := plugin.config
  stat, err := os.Stat(config["filename"])

  if err != nil {
    if os.IsNotExist(err) {
      syscall.Mkfifo(config["filename"], 0600)

      stat, _ = os.Stat(config["filename"])
    }
  }

  if (stat.Mode() & os.ModeNamedPipe) != os.ModeNamedPipe {
    log.Fatalf("[%T] %s is not a named pipe (stat.Mode=%s)", plugin, config["filename"], stat.Mode())
  }

  /* segfault here - dooh */
  pipe, err := os.OpenFile(config["filename"], os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModeNamedPipe)
  defer pipe.Close()

  if err != nil {
    log.Fatalf(fmt.Sprintf("[%T] Failed opening named pipe %s (%s)\n", plugin, config["filename"], err))
  }

  n, err := pipe.WriteString(fmt.Sprintf("%s\n", event.RawMessage))

  if err != nil {
    log.Printf(fmt.Sprintf("[%T] Failed writing to named pipe %s (%s)\n", plugin, config["filename"], err))
  }

  if plugin.debug {
    log.Printf("[%T] wrote %d bytes", plugin, n)
  }
}
