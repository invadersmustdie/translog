package translog

import (
  "bytes"
  "compress/zlib"
  "encoding/json"
  "fmt"
  "log"
  "net"
  "regexp"
  "strconv"
)

type GelfWriterPlugin struct {
  config          map[string]string
  socketWriter    NetworkSocketWriter
  conn            net.Conn
  peer            string
  compressMessage bool
  autoDiscovery   bool
  proto           string
  debug           bool
}

type GelfMessage struct {
  Event   *Event
  GelfStr string
}

var autodiscovery_patterns = make(map[string]*regexp.Regexp)

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

  plugin.autoDiscovery = false

  if len(config["autodiscovery"]) > 0 && config["autodiscovery"] == "true" {
    plugin.autoDiscovery = true
    plugin.precompileRegexpPatterns()
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

func (plugin *GelfWriterPlugin) precompileRegexpPatterns() {
  log.Printf("[%T] precompiling autodiscovery patterns", plugin)

  v, _ := regexp.Compile(`^\d+$`)
  autodiscovery_patterns["INTEGER"] = v
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

  ad := plugin.autoDiscovery

  for k, v := range event.Fields {
    if ad && autodiscovery_patterns["INTEGER"].MatchString(v) {
      val, _ := strconv.ParseInt(v, 10, 0)
      gelfFields[fmt.Sprintf("_%s", k)] = val
    } else {
      gelfFields[fmt.Sprintf("_%s", k)] = v
    }
  }

  jsonMsg, err := json.Marshal(gelfFields)

  if err != nil {
    log.Printf("[%T] ERROR: json encoding failed (%s)", plugin, err)
    return nil
  }

  msg.GelfStr = string(jsonMsg)

  return msg
}
