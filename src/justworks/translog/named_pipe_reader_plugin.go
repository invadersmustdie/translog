package translog

import (
  "bufio"
  "fmt"
  "io"
  "log"
  "os"
  "strconv"
  "strings"
  "time"
)

type NamedPipeReaderPlugin struct {
  config         map[string]string
  check_interval int64
  debug          bool
}

func (plugin *NamedPipeReaderPlugin) Configure(config map[string]string) {
  log.Printf("[%T] config %v", plugin, config)

  plugin.config = config

  if config["debug"] == "true" {
    plugin.debug = true
  }

  if len(config["source"]) < 1 {
    log.Fatalf("[%T] ERROR: missing configuration option 'source'", plugin)
  }

  if len(config["check_interval"]) > 0 {
    check_interval, err := strconv.ParseInt(config["check_interval"], 10, 0)

    if err != nil {
      log.Fatalf("[%T] failed reading check_interval. using fallback (%s)", plugin, err)
    }

    plugin.check_interval = check_interval
  } else {
    log.Printf("[%T] no option 'check_interval' set. using fallback '2'", plugin)
    plugin.check_interval = 2
  }
}

func (plugin *NamedPipeReaderPlugin) Start(c chan *Event) {
  config := plugin.config
  source := fmt.Sprintf("pipe://%s", config["source"])

  pipe, err := os.OpenFile(config["source"], os.O_RDONLY, 0600)

  if err != nil {
    log.Printf("[%T] failed to open named pipe %s", plugin, config["source"])
    return
  }

  stdin_buf := bufio.NewReader(pipe)

  for {
    line_in, hasMore, err := stdin_buf.ReadLine()

    if err == io.EOF {
      if plugin.debug {
        log.Printf("[%T] found EOF", plugin)
      }

      time.Sleep(time.Duration(plugin.check_interval) * time.Second)
      continue
    }

    if err != nil {
      log.Println(err)
      return
    }

    if hasMore {
      log.Println("[%T] Error: input line too long", plugin)
      return
    }

    str := strings.TrimSpace(string(line_in))
    if string(str) != "" {
      e := CreateEvent(source)
      e.SetRawMessage(str)

      c <- e
    }
  }
}
