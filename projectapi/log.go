package projectapi

import "github.com/uploadcare/uploadcare-go/v2/uclog"

var log uclog.Logger

const subsystemTag = "PRJA"

func init() { DisableLog() }

func DisableLog() {
	log = uclog.Disabled
}

func EnableLog(lvl uclog.Level) {
	log = uclog.Backend.Logger(subsystemTag)
	log.SetLevel(lvl)
}
