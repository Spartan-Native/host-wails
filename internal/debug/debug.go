// Package debug logs every call and change when SPARTAN_NATIVE_DEBUG is set.
package debug

import (
	"log"
	"os"
)

var enabled = os.Getenv("SPARTAN_NATIVE_DEBUG") != ""

// Logf logs with the [spartan-native] prefix when debugging is on.
func Logf(format string, args ...any) {
	if enabled {
		log.Printf("[spartan-native] "+format, args...)
	}
}
