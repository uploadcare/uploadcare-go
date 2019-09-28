package file

import (
	"github.com/uploadcare/uploadcare-go/uclog"
)

var log uclog.Logger

const subsystemTag = "FILE"

func init() { DisableLog() }

// DisableLog does what you expect
func DisableLog() { log = uclog.Disabled }

// EnableLog enables package scoped logging
func EnableLog(lvl uclog.Level) {
	log = uclog.Backend.Logger(subsystemTag)
	log.SetLevel(lvl)
}
