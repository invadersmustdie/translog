package translog

import (
  "bytes"
  "compress/zlib"
  "encoding/json"
  "fmt"
  "log"
  "net"
)

type GelfWriterPlugin struct {
  config          map[string]string
  socketWriter    NetworkSocketWriter
  conn            net.Conn
  peer            string
  compressMessage bool
  proto           string
  debug           bool
}

type GelfMessage struct {
  Event   *Event
  GelfStr string
}

func (plugin *GelfWriterPlugin) Configure(config map[string]string) {
  plugin.config = config

  if len(config["proto"]) > 0 && config["proto"] == "tcp" {
    log.Printf("[%T] GELF+TCP only supported uncompressed mode", plugin)
    plugin.compressMessage = false
  }

  if len(config["proto"]) > 0 && config["proto"] == "udp" {
    log.Printf("[%T] GELF+UDP enabled message compression", plugin)
    plugin.compressMessage = true
  }

  w := CreateNetworkSocketWriter(plugin, config)
  plugin.socketWriter = w

  if config["debug"] == "true" {
    plugin.debug = true
  }
}

func (plugin *GelfWriterPlugin) ProcessEvent(event *Event) {
  gelfMessage := plugin.CreateGelfMessage(event)

  if plugin.compressMessage {
    plugin.socketWriter.WriteBytes(plugin.CompressMessage(gelfMessage.GelfStr))
  } else {
    plugin.socketWriter.WriteString(gelfMessage.GelfStr)
  }
}

func (plugin *GelfWriterPlugin) CompressMessage(msg string) bytes.Buffer {
  var buf bytes.Buffer

  w := zlib.NewWriter(&buf)
  w.Write([]byte(msg))
  w.Close()

  return buf
}

func (plugin *GelfWriterPlugin) CreateGelfMessage(event *Event) *GelfMessage {
  msg := new(GelfMessage)
  msg.Event = event

  gelfFields := map[string]interface{}{
    "version":       "1.0",
    "host":          event.Host,
    "timestamp":     event.Time.Unix(),
    "short_message": event.RawMessage,
  }

  for k, v := range event.Fields {
    gelfFields[fmt.Sprintf("_%s", k)] = v
  }

  jsonMsg, err := json.Marshal(gelfFields)

  if err != nil {
    log.Printf("[%T] ERROR: json encoding failed (%s)", plugin, err)
    return nil
  }

  msg.GelfStr = string(jsonMsg)

  return msg
}
