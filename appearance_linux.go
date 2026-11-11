//go:build linux

package spartannative

import (
	"bufio"
	"os"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
)

// Linux: the XDG desktop portal (org.freedesktop.portal.Settings) over D-Bus,
// which GNOME, KDE and the other desktops implement. The desktop itself comes
// from XDG_CURRENT_DESKTOP.

const (
	portalDest    = "org.freedesktop.portal.Desktop"
	portalPath    = "/org/freedesktop/portal/desktop"
	portalSetting = "org.freedesktop.portal.Settings"
	nsAppearance  = "org.freedesktop.appearance"
	nsGnome       = "org.gnome.desktop.interface"
)

type linuxReader struct {
	once sync.Once
	conn *dbus.Conn
}

func newReader() reader { return &linuxReader{} }

func (r *linuxReader) platform() PlatformInfo {
	desktop := linuxDesktop(os.Getenv("XDG_CURRENT_DESKTOP"))
	var version *string
	if v := osRelease("VERSION_ID"); v != "" {
		version = strPtr(v)
	}
	return PlatformInfo{
		OS: "linux", Version: version, Desktop: &desktop,
		Skin: resolveSkin("linux", version, nil, &desktop), Confidence: "exact",
	}
}

func (r *linuxReader) appearance(skin string) AppearanceState {
	state := AppearanceState{Skin: skin, Accent: stockAccent(skin), Scheme: "light", Contrast: "normal", TextScale: 1}

	if v, ok := r.read(nsAppearance, "color-scheme"); ok {
		if n, ok := v.(uint32); ok && n == 1 { // 1 = prefer dark
			state.Scheme = "dark"
		}
	}
	if v, ok := r.read(nsAppearance, "accent-color"); ok {
		// (ddd) in sRGB 0–1; out-of-range values mean "no accent set"
		if rgb, ok := v.([]interface{}); ok && len(rgb) == 3 {
			c := make([]float64, 3)
			valid := true
			for i, x := range rgb {
				f, ok := x.(float64)
				valid = valid && ok && f >= 0 && f <= 1
				c[i] = f
			}
			if valid {
				state.Accent = rgbHex(c[0], c[1], c[2])
			}
		}
	}
	if v, ok := r.read(nsAppearance, "contrast"); ok {
		if n, ok := v.(uint32); ok && n == 1 { // 1 = higher contrast
			state.Contrast = "more"
		}
	}
	if v, ok := r.read(nsAppearance, "reduced-motion"); ok {
		if n, ok := v.(uint32); ok && n == 1 {
			state.ReduceMotion = true
		}
	}
	if v, ok := r.read(nsGnome, "text-scaling-factor"); ok {
		if f, ok := v.(float64); ok {
			state.TextScale = roundScale(f)
		}
	}
	return state
}

func (r *linuxReader) capabilities() map[string]CapabilityLevel {
	// nativeMenus is partial: context menus are native, but a frameless window (custom title
	// bar) doesn't show the window menu bar, so desktop-ui draws the menubar in the page.
	return map[string]CapabilityLevel{
		"systemAccent": Native, "colorScheme": Native, "highContrast": Native, "reduceTransparency": None,
		"textScale": Partial, "windowFocus": Native, "windowMaterial": None, "customTitleBar": Native,
		"nativeMenus": Partial, "notifications": Partial, "taskbarProgress": None, "fileDialogs": Native,
		"osVersion": Native, "linuxDesktop": Native,
	}
}

// read asks the portal for one setting. ReadOne (portal v2) first, then the older Read.
func (r *linuxReader) read(namespace, key string) (interface{}, bool) {
	r.once.Do(func() { r.conn, _ = dbus.ConnectSessionBus() })
	if r.conn == nil {
		return nil, false
	}
	obj := r.conn.Object(portalDest, portalPath)
	var v dbus.Variant
	if err := obj.Call(portalSetting+".ReadOne", 0, namespace, key).Store(&v); err == nil {
		return v.Value(), true
	}
	if err := obj.Call(portalSetting+".Read", 0, namespace, key).Store(&v); err == nil {
		if inner, ok := v.Value().(dbus.Variant); ok { // Read wraps the value twice
			return inner.Value(), true
		}
		return v.Value(), true
	}
	return nil, false
}

func osRelease(field string) string {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return ""
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		if k, v, ok := strings.Cut(s.Text(), "="); ok && k == field {
			return strings.Trim(v, `"`)
		}
	}
	return ""
}
