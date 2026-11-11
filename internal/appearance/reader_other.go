//go:build !windows && !darwin && !linux

package appearance

// Other platforms: stock values, everything reported as unavailable.

type otherReader struct{}

func New() Reader { return otherReader{} }

func (otherReader) Platform() Platform {
	return Platform{OS: "unknown", Skin: "web", Confidence: "unknown"}
}

func (otherReader) Appearance(skin string) Appearance {
	return Appearance{Skin: skin, Accent: stockAccent(skin), Scheme: "light", Contrast: "normal", TextScale: 1}
}

func (otherReader) Capabilities() map[string]CapabilityLevel {
	return capabilities(map[string]CapabilityLevel{"windowFocus": Native})
}
