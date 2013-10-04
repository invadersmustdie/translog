package main

import (
  "flag"
  "fmt"
  "justworks/translog"
  "log"
  "os"
  "runtime"
)

const APP_VERSION = "0.1.0"
const APP_NAME = "translog"

type Plugin interface {
  Configure(config map[string]string)
}

func CopyMap(src map[string]string) map[string]string {
  new_map := make(map[string]string)

  for k, v := range src {
    new_map[k] = v
  }

  return new_map
}

func createInputPlugin(name string) translog.InputPlugin {
  switch name {
  case "File":
    return new(translog.FileReaderPlugin)
  case "NamedPipe":
    return new(translog.NamedPipeReaderPlugin)
  case "Tcp":
    return new(translog.TcpReaderPlugin)
  }

  return nil
}

func createFilterPlugin(name string) translog.FilterPlugin {
  switch name {
  case "KeyValueExtractor":
    return new(translog.KeyValueExtractor)
  case "FieldExtractor":
    return new(translog.FieldExtractor)
  case "DropEventFilter":
    return new(translog.DropEventFilter)
  case "ModifyEventFilter":
    return new(translog.ModifyEventFilter)
  }

  return nil
}

func createOutputPlugin(name string) translog.ProcessingPlugin {
  switch name {
  case "Stdout":
    return new(translog.StdoutWriterPlugin)
  case "NamedPipe":
    return new(translog.NamedPipeWriterPlugin)
  case "Statsd":
    return new(translog.StatsdPlugin)
  case "Gelf":
    return new(translog.GelfWriterPlugin)
  case "NetworkSocket":
    return new(translog.NetworkSocketWriter)
  }

  return nil
}

func main() {
  flag.Usage = func() {
    fmt.Fprintf(os.Stderr, "Usage of %s (%s):\n", APP_NAME, APP_VERSION)
    flag.PrintDefaults()
  }

  var opt_configFile = flag.String("config", "main.conf", "Configuration")
  var opt_cpus = flag.Int("cpus", runtime.NumCPU(), "Number of CPUs to utilize")
  var opt_reportInterval = flag.Int("report_interval", 10, "Interval in seconds")
  var opt_logfile = flag.String("logfile", "", "Logfile")
  var opt_pidfile = flag.String("pidfile", "", "Pidfile")

  flag.Parse()

  if *opt_cpus > 0 {
    runtime.GOMAXPROCS(*opt_cpus)
  }

  if len(*opt_logfile) > 0 {
    log_file, err := os.OpenFile(*opt_logfile, os.O_WRONLY|os.O_CREATE, 0666)

    if err != nil {
      log.Fatal("Unable to open logfile %s (%s)", *opt_logfile, err)
    }

    log.SetOutput(log_file)
  }

  if len(*opt_pidfile) > 0 {
    pid_file, err := os.OpenFile(*opt_pidfile, os.O_WRONLY|os.O_CREATE, 0666)

    if err != nil {
      log.Fatal("Unable to write pidfile %s (%s)", *opt_pidfile, err)
    }

    pid_file.WriteString(fmt.Sprintf("%d\n", os.Getpid()))
    pid_file.Close()
  }

  // *** configure pipeline
  pipeline := new(translog.EventPipeline)
  pipeline.Init()
  pipeline.ReportInterval = *opt_reportInterval

  cfg := translog.ReadConfigurationFromFile(*opt_configFile)

  for _, input := range cfg.Inputs {
    plugin := createInputPlugin(input.Name)

    if plugin != nil {
      plugin.Configure(input.Config)
      pipeline.AddInput(plugin)
    } else {
      log.Printf("ERROR: No such plugin '%s' found", input.Name)
    }
  }

  for _, filter := range cfg.Filters {
    plugin := createFilterPlugin(filter.Name)

    if plugin != nil {
      plugin.Configure(filter.Config)
      pipeline.AddFilter(plugin)
    } else {
      log.Printf("ERROR: No such plugin '%s' found", filter.Name)
    }
  }

  for _, output := range cfg.Outputs {
    plugin := createOutputPlugin(output.Name)

    if plugin != nil {
      plugin.Configure(output.Config)
      pipeline.AddOutput(plugin)
    } else {
      log.Printf("ERROR: No such plugin '%s' found", output.Name)
    }
  }

  pipeline.Start()
}
