// Package appearance reads the OS appearance settings (accent, light/dark, contrast,
// transparency, motion, text size) and the OS version, and watches them for changes.
//
// There is one Reader per OS, picked at build time (reader_<os>.go).
package appearance

import "github.com/spartan-native/host-wails/internal/contract"

// Short names for the contract types the readers produce.
type (
	Platform        = contract.PlatformInfo
	Appearance      = contract.AppearanceState
	CapabilityLevel = contract.CapabilityLevel
)

const (
	Native  = contract.Native
	Partial = contract.Partial
	None    = contract.None
)

// Reader reads the OS. New returns the implementation for the current OS; tests use a fake.
type Reader interface {
	Platform() Platform
	Appearance(skin string) Appearance
	Capabilities() map[string]CapabilityLevel
}

// capabilities returns a level for every key in contract.Capabilities: the given ones, and
// None for the rest. It panics on a key the contract doesn't have, so a typo fails the tests.
func capabilities(levels map[string]CapabilityLevel) map[string]CapabilityLevel {
	out := make(map[string]CapabilityLevel, len(contract.Capabilities))
	for _, c := range contract.Capabilities {
		out[c] = None
	}
	for c, level := range levels {
		if _, known := out[c]; !known {
			panic("appearance: unknown capability " + c)
		}
		out[c] = level
	}
	return out
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
