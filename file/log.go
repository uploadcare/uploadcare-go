package file

import (
	"github.com/uploadcare/uploadcare-go/uclog"
)

var log uclog.Logger

const subsystemTag = "FILE"

func init() { DisableLog() }

func DisableLog() { log = uclog.Disabled }

func EnableLog(lvl uclog.Level) {
	log = uclog.Backend.Logger(subsystemTag)
	log.SetLevel(lvl)
}
