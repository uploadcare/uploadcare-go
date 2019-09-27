package ucare

import (
	"github.com/uploadcare/uploadcare-go/uclog"
)

var log uclog.Logger

const subsystemTag = "UCRE"

func init() { DisableLog() }

func DisableLog() { log = uclog.Disabled }

func EnableLog(lvl uclog.Level) {
	log = uclog.Backend.Logger(subsystemTag)
	log.SetLevel(lvl)
}
