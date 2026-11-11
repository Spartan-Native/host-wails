//go:build !windows && !darwin && !linux

package spartannative

// Other platforms: stock values, everything reported as unavailable.

type otherReader struct{}

func newReader() reader { return otherReader{} }

func (otherReader) platform() PlatformInfo {
	return PlatformInfo{OS: "unknown", Skin: "web", Confidence: "unknown"}
}

func (otherReader) appearance(skin string) AppearanceState {
	return AppearanceState{Skin: skin, Accent: stockAccent(skin), Scheme: "light", Contrast: "normal", TextScale: 1}
}

func (otherReader) capabilities() map[string]CapabilityLevel {
	caps := map[string]CapabilityLevel{}
	for _, c := range []string{"systemAccent", "colorScheme", "highContrast", "reduceTransparency", "textScale", "windowFocus",
		"windowMaterial", "customTitleBar", "nativeMenus", "notifications", "taskbarProgress", "fileDialogs", "osVersion", "linuxDesktop"} {
		caps[c] = None
	}
	caps["windowFocus"] = Native
	return caps
}
