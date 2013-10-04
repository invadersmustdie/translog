package main

import (
  "flag"
  "justworks/translog"
  "log"
  "os"
  "runtime"
)

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
  case "FileReaderPlugin":
    return new(translog.FileReaderPlugin)
  case "NamedPipeReaderPlugin":
    return new(translog.NamedPipeReaderPlugin)
  case "TcpReaderPlugin":
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
  case "StdoutWriterPlugin":
    return new(translog.StdoutWriterPlugin)
  case "NamedPipeWriterPlugin":
    return new(translog.NamedPipeWriterPlugin)
  case "StatsdPlugin":
    return new(translog.StatsdPlugin)
  case "GelfWriterPlugin":
    return new(translog.GelfWriterPlugin)
  case "NetworkSocketWriter":
    return new(translog.NetworkSocketWriter)
  }

  return nil
}

func main() {
  var opt_configFile = flag.String("config", "main.ini", "Configuration")
  var opt_cpus = flag.Int("cpus", runtime.NumCPU(), "Number of CPUs to utilize")
  var opt_reportInterval = flag.Int("report_interval", 10, "Interval in seconds")
  var opt_logfile = flag.String("logfile", "", "Logfile")

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
