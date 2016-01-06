package debug

import (
	"log"
)

var DebugFlag bool

func Debug(fmt string, args ...interface{}) {
	if DebugFlag {
		log.Printf(fmt, args...)
	}
}
