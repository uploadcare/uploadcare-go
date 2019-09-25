package uploadcare

import (
	"os"

	"github.com/btcsuite/btclog"
)

var log btclog.Logger

// Level is the level at which a logger is configured.
// All messages sent to a level which is below the current level are filtered.
type LogLevel uint32

// LogLevel constants.
const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError

	subsystemTag = "UCRE"
)

func init() { DisableLog() }

func DisableLog() { log = btclog.Disabled }

// UseLogger is used to enable logging in the context of the package
func EnableLog(lvl LogLevel) {
	// we don't care much about the log backend for now
	logBackend := btclog.NewBackend(os.Stderr)

	logger := logBackend.Logger(subsystemTag)
	logger.SetLevel(btclog.Level(lvl))

	log = logger
}
