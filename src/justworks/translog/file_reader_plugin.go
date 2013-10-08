package translog

import (
  "bytes"
  "fmt"
  "log"
  "os"
  "strings"
  "time"
)

type FileReaderPlugin struct {
  config          map[string]string
  debug           bool
  rescan_interval int
  _last_error_msg     string
  _last_error_cnt     int
}

const ERROR_OCCUR_LIMIT = 30

func (plugin *FileReaderPlugin) Configure(config map[string]string) {
  plugin.config = config

  if config["debug"] == "true" {
    plugin.debug = true
  }

  if len(config["source"]) < 1 {
    log.Fatalf("[%T] err='missing configuration option <source>'", plugin)
  }

  plugin.rescan_interval = 2
}

func (plugin *FileReaderPlugin) logError2(stage string, err string) {
  err_s := fmt.Sprintf("%s|%s", stage, err)

  if plugin._last_error_msg != err_s {
    plugin._last_error_msg = err_s
    plugin._last_error_cnt = 0
  }

  plugin._last_error_cnt += 1

  if plugin._last_error_cnt == 1 || plugin._last_error_cnt >= ERROR_OCCUR_LIMIT {
    log.Printf("2 [%T] stage='%s' source='%s' err='%s' occurence=%d",
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

func (plugin *FileReaderPlugin) Start(c chan *Event) {
  config := plugin.config
  source := fmt.Sprintf("file://%s", config["source"])

  for {
    file, err := os.OpenFile(config["source"], os.O_RDONLY, 0600)

    if err != nil {
      plugin.logError2("open", err.Error())
      time.Sleep(time.Duration(plugin.rescan_interval) * time.Second)
      continue
    }

    stat, err := os.Stat(config["source"])

    if err != nil {
      plugin.logError2("statAfterOpen", err.Error())
      time.Sleep(time.Duration(plugin.rescan_interval) * time.Second)
      continue
    }

    prev_file_size := stat.Size()

    // jump to end of file
    file.Seek(0, os.SEEK_END)

    buf := bytes.NewBufferString("")

    for {
      stat, err := os.Stat(config["source"])

      if err != nil {
        plugin.logError2("read", err.Error())
        time.Sleep(time.Duration(plugin.rescan_interval) * time.Second)
        break
      }

      new_bytes := stat.Size() - prev_file_size

      if plugin.debug {
        log.Printf("[%T] polling %s", plugin, config["source"])
      }

      if new_bytes < 0 {
        log.Printf("[%T] rewinding file %s", plugin, config["source"])
        file.Seek(0, os.SEEK_SET)
      } else {
        if new_bytes != 0 {
          rbuf := make([]byte, new_bytes)
          _, err := file.Read(rbuf)

          if err != nil {
            plugin.logError2("readbuf", err.Error())
            time.Sleep(time.Duration(plugin.rescan_interval) * time.Second)
            continue
          }

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
        } else {
          if plugin.debug {
            log.Printf("[%T] %d pending bytes in buf", plugin, buf.Len())
            log.Printf("[%T] pending -> %s", plugin, buf.String())
          }

          if buf.Len() > 0 {
            if plugin.debug {
              log.Printf("[%T] Flushing pending buffer", plugin)
            }

            event := CreateEvent(source)
            event.SetRawMessage(buf.String())

            buf.Reset()
            c <- event
          }
        }
      }

      prev_file_size = stat.Size()
      stat, err = os.Stat(config["source"])

      if err != nil {
        plugin.logError2("stat2", err.Error())
        time.Sleep(time.Duration(plugin.rescan_interval) * time.Second)
        continue
      }

      // nothing has changed
      if stat.Size() == prev_file_size {
        time.Sleep(time.Duration(plugin.rescan_interval) * time.Second)
      }
    }
  }
}
