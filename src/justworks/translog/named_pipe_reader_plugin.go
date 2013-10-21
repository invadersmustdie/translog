package translog

import (
  "bytes"
  "fmt"
  "io"
  "log"
  "os"
  "strconv"
  "strings"
)

type NamedPipeReaderPlugin struct {
  config          map[string]string
  debug           bool
  bufsize         int
  _last_error_msg string
  _last_error_cnt int
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

  plugin.bufsize = 1024
  if len(config["bufsize"]) > 0 {
    val, err := strconv.ParseInt(config["bufsize"], 10, 0)

    if err != nil {
      log.Fatalf("[%T] failed reading bufsize. using fallback (%s)", plugin, 1024)
    }

    plugin.bufsize = int(val)
  }
}

// TODO: extract into common function
func (plugin *NamedPipeReaderPlugin) logError2(stage string, err string) {
  err_s := fmt.Sprintf("%s|%s", stage, err)

  if plugin._last_error_msg != err_s {
    plugin._last_error_msg = err_s
    plugin._last_error_cnt = 0
  }

  plugin._last_error_cnt += 1

  if plugin._last_error_cnt == 1 || plugin._last_error_cnt >= ERROR_OCCUR_LIMIT {
    log.Printf("[%T] stage='%s' source='%s' err='%s' occurence=%d",
      plugin,
      stage,
      plugin.config["source"],
      err,
      plugin._last_error_cnt)
  }

  if plugin._last_error_cnt >= ERROR_OCCUR_LIMIT {
    plugin._last_error_cnt = 1
  }
}

func (plugin *NamedPipeReaderPlugin) Start(c chan *Event) {
  config := plugin.config
  source := fmt.Sprintf("pipe://%s", config["source"])

  pipe, err := os.OpenFile(config["source"], os.O_RDONLY, 0600)

  if err != nil {
    plugin.logError2("openFile", err.Error())
    return
  }

  buf := bytes.NewBufferString("")

  for {
    rbuf := make([]byte, plugin.bufsize)
    _, err = pipe.Read(rbuf)

    if err != io.EOF {
      for _, char := range rbuf {
        if char != '\n' {
          buf.WriteByte(char)
        } else {
          if len(strings.TrimSpace(buf.String())) > 0 {
            event := CreateEvent(source)
            event.SetRawMessage(buf.String())

            c <- event
            buf.Reset()
          }
        }
      }

      if err != nil {
        plugin.logError2("write", err.Error())
      }
    }
  }
}
