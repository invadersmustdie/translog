package translog

import (
  "bufio"
  "fmt"
  "log"
  "net"
)

type NetworkSocketWriter struct {
  config                   map[string]string
  conn                     net.Conn
  usePersistentConnections bool
  peer                     string
  proto                    string
  Caller                   string
  debug                    bool
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
}

func (plugin *NetworkSocketWriter) ProcessEvent(event *Event) {
  plugin.WriteString(event.RawMessage)
}

func (plugin *NetworkSocketWriter) WriteString(data string) {
  if plugin.usePersistentConnections {
    if plugin.debug {
      log.Printf("[%s > %T] Using persistent connection", plugin.Caller, plugin)
    }

    /* NOTE: still work in progress */
    plugin.WriteStringToConnnectionPool(data)
  } else {
    if plugin.debug {
      log.Printf("[%s > %T] Using single shot connection", plugin.Caller, plugin)
    }

    plugin.WriteStringSingleShot(data)
  }
}

func (plugin *NetworkSocketWriter) WriteStringSingleShot(data string) {
  if plugin.debug {
    log.Printf("[%s > %T] Establishing connection to %s", plugin.Caller, plugin, plugin.peer)
  }

  conn, err := net.Dial(plugin.proto, plugin.peer)

  if err != nil {
    log.Printf("[%s > %T] Failed opening %s://%s (%s)", plugin.Caller, plugin, plugin.proto, plugin.peer, err)
    return
  }

  if plugin.debug {
    log.Printf("[%s > %T] Writing to '%s' %s", plugin.Caller, plugin, data, plugin.peer)
  }

  writer := bufio.NewWriter(conn)
  writer.WriteString(data)
  writer.Flush()

  defer conn.Close()
}

func (plugin *NetworkSocketWriter) WriteStringToConnnectionPool(data string) {
  if plugin.conn == nil {
    if plugin.debug {
      log.Printf("[%s > %T] Establishing connection to %s", plugin.Caller, plugin, plugin.peer)
    }

    conn, err := net.Dial(plugin.proto, plugin.peer)
    plugin.conn = conn

    if err != nil {
      log.Printf("[%s > %T] Failed opening %s://%s (%s)", plugin.Caller, plugin, plugin.proto, plugin.peer, err)
    }
  } else {
    if plugin.debug {
      log.Printf("[%s > %T] Reusing connection %s://%s", plugin.Caller, plugin, plugin.proto, plugin.peer)
    }
  }

  if plugin.conn != nil {
    if plugin.debug {
      log.Printf("[%s > %T] Writing to '%s' %s", plugin.Caller, plugin, data, plugin.peer)
    }

    writer := bufio.NewWriter(plugin.conn)
    writer.WriteString(data)
    writer.Flush()
  } else {
    log.Printf("[%s > %T] ERROR: no available connection found", plugin.Caller, plugin)
  }
}
