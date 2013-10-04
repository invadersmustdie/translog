package translog

import (
  "log"
)

/* http://play.golang.org/p/QFheQeChIn */
const debug debugging = false

type debugging bool

func (d debugging) Printf(format string, args ...interface{}) {
  if d {
    log.Printf(format, args...)
  }
}
