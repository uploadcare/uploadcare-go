// Package uclog holds basic leveled logging primitives. Basically, it just makes
// dependancy tree cleaner
package uclog

import (
	"os"

	"github.com/btcsuite/btclog"
)

// Logger is an interface which describes a level-based logger.
type Logger btclog.Logger

// Level is the level at which a logger is configured.
// All messages sent to a level which is below the current level are filtered.
type Level = btclog.Level

// LogLevel constants (basic list)
const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var (
	// Backend is a default log backend
	// TODO: make it settable
	Backend = btclog.NewBackend(os.Stderr)

	// Disabled is a Logger that will never output anything.
	Disabled = btclog.Disabled
)
