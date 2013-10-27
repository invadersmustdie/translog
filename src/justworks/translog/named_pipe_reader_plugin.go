package translog

import (
  "bytes"
  "fmt"
  "io"
  "log"
  "os"
  "strconv"
  "strings"
  "time"
  "syscall"
)

type NamedPipeReaderPlugin struct {
  config          map[string]string
  debug           bool
  bufsize         int
  poll_interval   time.Duration
  check_interval  int
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

  plugin.poll_interval = time.Duration(1) * time.Second
  plugin.check_interval = 10
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

func (plugin *NamedPipeReaderPlugin) openPipe() (*os.File, os.FileInfo, error) {
  pipe, err := os.OpenFile(plugin.config["source"], os.O_RDONLY, 0600)

  if err != nil {
    plugin.logError2("openFile", err.Error())
    return pipe, nil, err
  }

  pipeStat, err := pipe.Stat()

  if err != nil {
    plugin.logError2("pipeStat", err.Error())
    return nil, nil, err
  }

  if err = syscall.SetNonblock(int(pipe.Fd()), true); err != nil {
    log.Printf("[%T] Failed opening file in NONBLOCKING mode (err='%s')", err.Error())
  }

  return pipe, pipeStat, nil
}

func (plugin *NamedPipeReaderPlugin) Start(c chan *Event) {
  config := plugin.config
  source := fmt.Sprintf("pipe://%s", config["source"])

  pipe, pipeStat, err := plugin.openPipe()

  if err != nil {
    log.Fatalf("[%T] failed to open named pipe (err='%s')", err.Error())
  }

  buf := bytes.NewBufferString("")

  check_loop_cnt := 0
  do_read := true

  for {
    if check_loop_cnt >= plugin.check_interval {
      if plugin.debug {
        log.Printf("[%T] checking for changes on named pipe %s", plugin, config["source"])
      }

      fi, err := os.Stat(config["source"])

      if err != nil {
        if do_read {
          plugin.logError2("checkPipe1", err.Error())
        }
        do_read = false
      }

      if !os.SameFile(pipeStat, fi) {
        if plugin.debug {
          log.Printf("[%T] detected change in named pipe, reloading %s", plugin, config["source"])
        }

        pipe, pipeStat, err = plugin.openPipe()

        if err != nil {
          if do_read {
            plugin.logError2("openNewPipe", err.Error())
          }

          do_read = false
        } else {
          log.Printf("[%T] detected change in named pipe, reloaded %s", plugin, config["source"])
          do_read = true
        }
      } else {
        do_read = true
      }

      check_loop_cnt = 0
    }

    check_loop_cnt += 1

    if do_read {
      rbuf := make([]byte, plugin.bufsize)
      n, err := pipe.Read(rbuf)

      if n != 0 && err == nil {
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
      }

      if n == 0 && err != nil && err != io.EOF {
        patherr := err.(*os.PathError)

        if !(patherr != nil && patherr.Err.Error() == "resource temporarily unavailable") {
          plugin.logError2("write", err.Error())
        }
      }

      if n == 0 || err == io.EOF {
        time.Sleep(plugin.poll_interval)
      }
    } else {
      time.Sleep(plugin.poll_interval)
    }
  }
}
