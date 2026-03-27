package projectapi

import "github.com/uploadcare/uploadcare-go/v2/uclog"

var log uclog.Logger

const subsystemTag = "PRJA"

func init() {
	DisableLog()
}

// DisableLog disables all log output for this package
func DisableLog() {
	log = uclog.Disabled
}

// EnableLog enables log output for this package at the given level
func EnableLog(lvl uclog.Level) {
	log = uclog.Backend.Logger(subsystemTag)
	log.SetLevel(lvl)
}
