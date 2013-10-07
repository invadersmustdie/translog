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
}

const ERROR_OCCUR_LIMIT = 100

func (plugin *FileReaderPlugin) Configure(config map[string]string) {
  plugin.config = config

  if config["debug"] == "true" {
    plugin.debug = true
  }

  if len(config["source"]) < 1 {
    log.Fatalf("[%T] ERROR: missing configuration option 'source'", plugin)
  }

  plugin.rescan_interval = 2
}

func (plugin *FileReaderPlugin) Start(c chan *Event) {
  config := plugin.config
  source := fmt.Sprintf("file://%s", config["source"])

  error_open_count := 0
  error_read_count := 0

  for {
    file, err := os.OpenFile(config["source"], os.O_RDONLY, 0600)

    if err != nil {
      error_open_count += 1

      if error_open_count == 1 || error_open_count >= ERROR_OCCUR_LIMIT {
        log.Printf("[%T] ERROR: failed to open file %s (error='%s') (this message occured %d times)", plugin, config["source"], err, error_open_count)
      }

      if error_open_count >= ERROR_OCCUR_LIMIT {
        error_open_count = 1
      }

      time.Sleep(time.Duration(plugin.rescan_interval) * time.Second)
      continue
    }

    stat, _ := os.Stat(config["source"])
    prev_file_size := stat.Size()

    // jump to end of file
    file.Seek(0, os.SEEK_END)

    buf := bytes.NewBufferString("")

    for {
      stat, _ := os.Stat(config["source"])
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
          n, err := file.Read(rbuf)

          if err != nil {
            error_read_count += 1

            if error_read_count == 1 || error_read_count >= ERROR_OCCUR_LIMIT {
              log.Printf("[%T] ERROR: failed reading %s (new_bytes=%d, n=%d) error='%s' (this message occured %d times)", plugin, config["source"], new_bytes, n, err, error_read_count)
            }

            if error_read_count >= ERROR_OCCUR_LIMIT {
              error_read_count = 1
            }

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
      stat, _ = os.Stat(config["source"])

      if stat.Size() == prev_file_size {
        time.Sleep(time.Duration(plugin.rescan_interval) * time.Second)
      }
    }
  }
}
