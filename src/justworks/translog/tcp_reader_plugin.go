package translog

import (
  "bytes"
  "fmt"
  "io"
  "log"
  "net"
  "strconv"
  "time"
)

type TcpReaderPlugin struct {
  config        map[string]string
  sleepDuration time.Duration
}

func (plugin *TcpReaderPlugin) Configure(config map[string]string) {
  config["proto"] = "tcp"
  config["sleep_interval"] = "2"

  if len(config["port"]) > 0 {
    config["port"] = fmt.Sprintf(":%s", config["port"])
  } else {
    config["port"] = ":9999"
  }

  sleep_duration, err := strconv.ParseInt(config["sleep_interval"], 10, 0)

  if err != nil {
    log.Printf("Failed parsing sleep_interval")
    plugin.sleepDuration = 2 * time.Second
  } else {
    plugin.sleepDuration = time.Duration(sleep_duration) * time.Second
  }

  plugin.config = config
}

func (plugin *TcpReaderPlugin) Start(c chan *Event) {
  config := plugin.config

  log.Printf("[%T] starting tcp listener on %s", plugin, config["port"])
  ln, err := net.Listen(config["proto"], config["port"])
  conn_counter := 0

  if err != nil {
    log.Printf("ERROR: error while listening on %s:%s (%s)", config["proto"], config["port"], err)
    return
  } else {
    for {
      conn, err := ln.Accept()
      conn_counter += 1

      if err != nil {
        log.Printf("ERROR: error while accepting on %s:%s (%s)", config["proto"], config["port"], err)
        debug.Printf("DEBUG: sleeping for %d seconds", plugin.sleepDuration)

        time.Sleep(plugin.sleepDuration)
      }

      defer conn.Close()
      log.Printf("[TcpReaderPlugin] new incoming connection (%s)", conn.RemoteAddr().String())
      go plugin.handleConnection(conn, c)
    }
  }
}

func (plugin *TcpReaderPlugin) handleConnection(conn net.Conn, c chan *Event) {
  buf := bytes.NewBufferString("")
  source := fmt.Sprintf("%s://localhost%s", plugin.config["proto"], plugin.config["port"])

  for {
    data := make([]byte, 256)
    _, err := conn.Read(data)

    if err != nil {
      if err == io.EOF {
        log.Printf("[TcpReaderPlugin] peer (%s) closed connection", conn.RemoteAddr().String())
      } else {
        log.Printf("ERROR: failed reading data from tcp socket. (%s)\n", err)
      }
      return
    }

    for _, char := range data {
      if char != '\n' {
        buf.WriteByte(char)
      } else {
        event := CreateEvent(source)
        event.SetRawMessage(buf.String())

        debug.Printf("data -> %s\n", buf.String())

        buf.Reset()
        c <- event
      }
    }

  }
}
