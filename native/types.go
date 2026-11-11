// Package spartannative is the native side of @spartan-native/host-wails.
//
// Register the service with your Wails v3 application:
//
//	app := application.New(application.Options{
//		Services: []application.Service{
//			application.NewService(spartannative.New()),
//		},
//	})
//
// The frontend calls Platform(), Appearance() and Capabilities() through the
// Wails bindings, and receives the "spartan-native:appearance" event whenever
// the OS appearance changes (accent, light/dark, contrast, transparency,
// motion, text size).
package spartannative

// EventAppearance is emitted with an AppearanceState whenever a value changes.
const EventAppearance = "spartan-native:appearance"

// PlatformInfo mirrors PlatformInfo in @spartan-native/theme.
type PlatformInfo struct {
	OS         string  `json:"os"`      // windows | macos | linux | chromeos | unknown
	Version    *string `json:"version"` // "10.0.22631", "26.1", …
	Build      *int    `json:"build"`   // Windows build; ≥ 22000 means Windows 11
	Desktop    *string `json:"desktop"` // gnome | kde | cinnamon | xfce | other (Linux)
	Skin       string  `json:"skin"`    // win | w10 | mac | macc | gtk | qt | cros | web
	Confidence string  `json:"confidence"`
}

// AppearanceState mirrors AppearanceState in @spartan-native/theme.
type AppearanceState struct {
	Skin               string  `json:"skin"`
	Accent             string  `json:"accent"`   // #rrggbb
	Scheme             string  `json:"scheme"`   // light | dark
	Contrast           string  `json:"contrast"` // normal | more | forced
	ReduceTransparency bool    `json:"reduceTransparency"`
	ReduceMotion       bool    `json:"reduceMotion"`
	TextScale          float64 `json:"textScale"` // 1 = 100 %
}

// CapabilityLevel is native, partial or none.
type CapabilityLevel = string

const (
	Native  CapabilityLevel = "native"
	Partial CapabilityLevel = "partial"
	None    CapabilityLevel = "none"
)

// reader reads the OS. One implementation per OS (appearance_<os>.go); tests use a fake.
type reader interface {
	platform() PlatformInfo
	appearance(skin string) AppearanceState
	capabilities() map[string]CapabilityLevel
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
