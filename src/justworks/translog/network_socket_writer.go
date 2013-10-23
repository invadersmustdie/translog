package translog

import (
  "bufio"
  "bytes"
  "fmt"
  "log"
  "net"
  "strconv"
  "time"
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
  plugin.WriteString(fmt.Sprintf("%s\r\n", event.RawMessage))
}

func (plugin *NetworkSocketWriter) initializePool() {
  if plugin.pool != nil {
    return
  }

  dial_timeout := 1 * time.Second
  plugin.pool = make(chan *net.Conn, plugin.pool_size)

  for i := 0; i <= plugin.pool_size; i++ {
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
    log.Printf("[%s > %T] getConnection", plugin.Caller, plugin)
  }

  return <-plugin.pool
}

func (plugin *NetworkSocketWriter) releaseConnection(conn *net.Conn) {
  if plugin.debug {
    log.Printf("[%s > %T] releaseConnection", plugin.Caller, plugin)
  }

  plugin.pool <- conn
}

func (plugin *NetworkSocketWriter) WriteString(data string) {
  conn := plugin.getConnection()
  defer plugin.releaseConnection(conn)

  if plugin.debug {
    log.Printf("[%s > %T] conn=%s WriteString('%s')", plugin.Caller, plugin, conn, data)
  }

  writer := bufio.NewWriter(*conn)
  writer.WriteString(data)
  writer.Flush()
}

func (plugin *NetworkSocketWriter) WriteBytes(data bytes.Buffer) {
  conn := plugin.getConnection()
  defer plugin.releaseConnection(conn)

  if plugin.debug {
    log.Printf("[%s > %T] conn=%s WriteBytes size=%d string='%x'",
      plugin.Caller,
      plugin,
      conn,
      data.Len(),
      data.Bytes())
  }

  n, err := (*conn).Write(data.Bytes())

  if err != nil {
    log.Printf("[%s > %T] WriteBytes failed (err=%s)", plugin.Caller, plugin, err)
  }

  if plugin.debug {
    log.Printf("[%s > %T] WriteBytes completed (size=%d)", plugin.Caller, plugin, n)
  }
}
