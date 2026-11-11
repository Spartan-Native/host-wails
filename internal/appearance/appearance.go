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

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
