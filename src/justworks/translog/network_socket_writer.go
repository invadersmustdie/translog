package translog

import (
  "bufio"
  "fmt"
  "log"
  "net"
  "bytes"
  "time"
  "strconv"
)

type NetworkSocketWriter struct {
  config                   map[string]string
  conn                     net.Conn
  usePersistentConnections bool
  peer                     string
  proto                    string
  Caller                   string
  debug                    bool
  pool                     chan *net.Conn
  pool_size                int
}

func CreateNetworkSocketWriter(caller interface{}, config map[string]string) NetworkSocketWriter {
  w := new(NetworkSocketWriter)
  w.Caller = fmt.Sprintf("%T", caller)
  w.Configure(config)

  return *w
}

func (plugin *NetworkSocketWriter) Configure(config map[string]string) {
  plugin.config = config

  if config["debug"] == "true" {
    plugin.debug = true
  }

  if len(config["host"]) > 0 && len(config["port"]) > 0 {
    plugin.peer = fmt.Sprintf("%s:%s", config["host"], config["port"])
  } else {
    log.Fatalf("[%s > %T] Invalid value for option 'host' and/or 'port' set", plugin.Caller, plugin)
  }

  if len(config["proto"]) == 0 {
    log.Fatalf("[%s > %T] Missing option 'proto'", plugin.Caller, plugin)
  } else {
    plugin.proto = config["proto"]
  }

  plugin.pool_size = 5
  if len(config["pool_size"]) > 0 {
    val, err := strconv.ParseInt(config["pool_size"], 10, 0)

    if err != nil {
      log.Fatalf("[%s > %T] Invalid value for option 'pool_size'", plugin.Caller, plugin)
    }

    plugin.pool_size = int(val)
  }
}

func (plugin *NetworkSocketWriter) ProcessEvent(event *Event) {
  plugin.WriteString(event.RawMessage)
}

func (plugin *NetworkSocketWriter) initializePool() {
  if plugin.pool != nil {
    return
  }

  dial_timeout := 1 * time.Second
  plugin.pool = make(chan *net.Conn, plugin.pool_size)

  for i := 0; i<= plugin.pool_size; i++ {
    conn, err := net.DialTimeout(plugin.proto, plugin.peer, dial_timeout)

    if err != nil {
      log.Printf("[%s > %T] Failed opening %s://%s (%s)", plugin.Caller, plugin, plugin.proto, plugin.peer, err)
    } else {
      plugin.pool <- &conn
    }
  }
}

func (plugin *NetworkSocketWriter) getConnection() *net.Conn {
  plugin.initializePool()

  if plugin.debug {
    log.Printf("[%T] getConnection", plugin)
  }

  return <- plugin.pool
}

func (plugin *NetworkSocketWriter) releaseConnection(conn *net.Conn) {
  if plugin.debug {
    log.Printf("[%T] releaseConnection", plugin)
  }

  plugin.pool <- conn
}

func (plugin *NetworkSocketWriter) WriteString(data string) {
  conn := plugin.getConnection()
  defer plugin.releaseConnection(conn)

  if plugin.debug {
    log.Printf("[%T] conn=%s WriteString('%s')", plugin, conn, data)
  }

  writer := bufio.NewWriter(*conn)
  writer.WriteString(data)
  writer.Flush()
}

func (plugin *NetworkSocketWriter) WriteBytes(data bytes.Buffer) {
  conn := plugin.getConnection()
  defer plugin.releaseConnection(conn)

  if plugin.debug {
    log.Printf("[%T] conn=%s WriteBytes('%s')", plugin, conn, data.String())
  }

  (*conn).Write(data.Bytes())
}
