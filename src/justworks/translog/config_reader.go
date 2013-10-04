package translog

import (
  "bufio"
  "fmt"
  "io"
  "log"
  "os"
  "regexp"
  "strings"
)

type Configuration struct {
  Inputs  []*Section
  Filters []*Section
  Outputs []*Section
}

type Section struct {
  Name   string
  T      string
  Config map[string]string
}

func ReadConfigurationFromString(config string) Configuration {
  reader := bufio.NewReader(strings.NewReader(config))

  return readConfiguration(*reader)
}

func ReadConfigurationFromFile(filename string) Configuration {
  file, err := os.OpenFile(filename, os.O_RDONLY, 0600)

  if err != nil {
    log.Fatal(fmt.Sprintf("Failed to open configuration %s (%s)", filename, err))
  }

  reader := bufio.NewReader(file)

  return readConfiguration(*reader)
}

func readConfiguration(reader bufio.Reader) Configuration {
  cfg := new(Configuration)

  section_re, _ := regexp.Compile(`^\[(\w+)\.(\w+)`)
  key_value_re, _ := regexp.Compile(`^(\w+(?:\.\w+)*)\s*=\s*(.+)$`)

  section := new(Section)

  for {
    line, _, err := reader.ReadLine()

    if section_re.MatchString(string(line)) || err == io.EOF {
      if len(section.Name) > 0 {
        if section.T == "input" {
          cfg.Inputs = append(cfg.Inputs, section)
        }

        if section.T == "filter" {
          cfg.Filters = append(cfg.Filters, section)
        }

        if section.T == "output" {
          cfg.Outputs = append(cfg.Outputs, section)
        }

        if err == io.EOF {
          break
        }
      }

      result := section_re.FindAllStringSubmatch(string(line), -1)

      if len(result) > 0 {
        section = new(Section)
        section.Name = result[0][2]
        section.T = result[0][1]
        section.Config = make(map[string]string)
      }

      continue
    }

    if key_value_re.MatchString(string(line)) {
      result := key_value_re.FindAllStringSubmatch(string(line), -1)

      if len(section.Name) > 0 && len(result) > 0 {
        section.Config[result[0][1]] = result[0][2]
      }

      continue
    }
  }

  return *cfg
}
