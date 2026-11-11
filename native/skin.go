package spartannative

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// resolveSkin mirrors resolveSkin() in @spartan-native/theme.
func resolveSkin(os string, version *string, build *int, desktop *string) string {
	switch os {
	case "windows":
		if build != nil && *build < 22000 {
			return "w10"
		}
		return "win"
	case "macos":
		if version != nil {
			if major, err := strconv.Atoi(strings.SplitN(*version, ".", 2)[0]); err == nil && major < 26 {
				return "macc"
			}
		}
		return "mac"
	case "linux":
		if desktop != nil && *desktop == "kde" {
			return "qt"
		}
		return "gtk"
	case "chromeos":
		return "cros"
	}
	return "web"
}

// linuxDesktop maps XDG_CURRENT_DESKTOP ("ubuntu:GNOME", "KDE", …) to a desktop name.
func linuxDesktop(xdg string) string {
	v := strings.ToLower(xdg)
	switch {
	case strings.Contains(v, "kde"):
		return "kde"
	case strings.Contains(v, "gnome"), strings.Contains(v, "unity"), strings.Contains(v, "budgie"), strings.Contains(v, "pantheon"):
		return "gnome"
	case strings.Contains(v, "cinnamon"), strings.Contains(v, "x-cinnamon"):
		return "cinnamon"
	case strings.Contains(v, "xfce"):
		return "xfce"
	}
	return "other"
}

// rgbHex formats 0–1 color components as #rrggbb.
func rgbHex(r, g, b float64) string {
	c := func(v float64) int { return int(math.Round(math.Max(0, math.Min(1, v)) * 255)) }
	return fmt.Sprintf("#%02x%02x%02x", c(r), c(g), c(b))
}

// stockAccent is used when the OS has no accent setting.
func stockAccent(skin string) string {
	switch skin {
	case "mac", "macc":
		return "#007aff"
	case "gtk":
		return "#3584e4"
	case "qt":
		return "#3daee9"
	}
	return "#0078d4"
}

// roundScale keeps text scales tidy (0.05 steps).
func roundScale(v float64) float64 {
	if v <= 0 {
		return 1
	}
	return math.Round(v*20) / 20
}
