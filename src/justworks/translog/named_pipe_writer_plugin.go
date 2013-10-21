package translog

import (
  "fmt"
  "log"
  "os"
  "strconv"
  "sync"
  "sync/atomic"
  "syscall"
  "time"
)

type NamedPipeWriterPlugin struct {
  config       map[string]string
  debug        bool
  openFileLock sync.Mutex
  open_msgs    int64
  max_messages int64
}

const NAMED_PIPE_WRITER_PLUGIN_MAX_MESSAGES = 5000

func (plugin *NamedPipeWriterPlugin) Configure(config map[string]string) {
  log.Printf("[%T] config %v", plugin, config)

  plugin.config = config

  if config["debug"] == "true" {
    plugin.debug = true
  }

  if len(config["filename"]) < 1 {
    log.Fatalf("[%T] ERROR: missing configuration option 'filename'", plugin)
  }

  if len(config["max_messages"]) > 0 {
    val, err := strconv.ParseInt(config["max_messages"], 10, 0)

    if err != nil {
      log.Fatalf("[%T] ERROR: invalid value for max_messages(int) set (err=%s)", plugin, err)
    }

    plugin.max_messages = val
  }

  if plugin.max_messages <= 0 {
    plugin.max_messages = NAMED_PIPE_WRITER_PLUGIN_MAX_MESSAGES
  }

  report_interval := 30
  if len(config["report_interval"]) > 0 {
    val, err := strconv.ParseInt(config["report_interval"], 10, 0)

    if err != nil {
      log.Fatalf("[%T] ERROR: invalid value for report_interval(int) set (err=%s)", plugin, err)
    }

    report_interval = int(val)
  }

  go plugin.pipeMonitor(report_interval)
}

func (plugin *NamedPipeWriterPlugin) pipeMonitor(sleepInterval int) {
  for {
    if atomic.LoadInt64(&plugin.open_msgs) >= plugin.max_messages {
      log.Printf("[%T] WARNING: filling level reached maximum (max=%d) - incoming messages will be dropped",
        plugin,
        plugin.max_messages)
    }

    if plugin.debug {
      log.Printf("[%T] stats: watermark=%d max=%d",
        plugin,
        atomic.LoadInt64(&plugin.open_msgs),
        plugin.max_messages)
    }

    time.Sleep(time.Duration(sleepInterval) * time.Second)
  }
}

func (plugin *NamedPipeWriterPlugin) ProcessEvent(event *Event) {
  config := plugin.config
  var pipe *os.File

  if atomic.LoadInt64(&plugin.open_msgs) >= plugin.max_messages {
    return
  }

  atomic.AddInt64(&plugin.open_msgs, 1)
  plugin.openFileLock.Lock()

  stat, err := os.Stat(config["filename"])

  if err != nil {
    if os.IsNotExist(err) {
      log.Printf("[%T] creating named pipe %s", plugin, config["filename"])
      syscall.Mkfifo(config["filename"], 0600)

      stat, err = os.Stat(config["filename"])

      if err != nil {
        log.Fatalf("[%T] stat failed (err=%s)", plugin, err)
      }
    } else {
      log.Fatalf("[%T] failed opening %s (err=%s)", plugin, config["filename"], err)
    }
  }

  pipe, err = os.OpenFile(config["filename"], os.O_WRONLY, 0600)

  if (stat.Mode() & os.ModeNamedPipe) != os.ModeNamedPipe {
    log.Fatalf("[%T] %s is not a named pipe (stat.Mode=%s)", plugin, config["filename"], stat.Mode())
  }

  plugin.openFileLock.Unlock()

  n, err := pipe.WriteString(fmt.Sprintf("%s\n", event.RawMessage))

  if err != nil {
    log.Printf(fmt.Sprintf("[%T] Failed writing to named pipe %s (%s)\n", plugin, config["filename"], err))
    return
  }

  if plugin.debug {
    log.Printf("[%T] wrote %d bytes", plugin, n)
  }

  atomic.AddInt64(&plugin.open_msgs, -1)
}
