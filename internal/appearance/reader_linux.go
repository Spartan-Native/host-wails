//go:build linux

package appearance

import (
	"bufio"
	"context"
	"os"
	"strings"
	"sync"
	"time"

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

	// readTimeout bounds one Appearance() read, all settings together, so a hung portal
	// can't block the frontend's call for D-Bus' default 25 seconds.
	readTimeout = 2 * time.Second
	// reconnectEvery limits how often a missing session bus is retried.
	reconnectEvery = 5 * time.Second
)

type linuxReader struct {
	mu          sync.Mutex
	conn        *dbus.Conn
	lastAttempt time.Time
}

func New() Reader { return &linuxReader{} }

func (r *linuxReader) Platform() Platform {
	desktop := linuxDesktop(os.Getenv("XDG_CURRENT_DESKTOP"))
	var version *string
	if v := osRelease("VERSION_ID"); v != "" {
		version = strPtr(v)
	}
	return Platform{
		OS: "linux", Version: version, Desktop: &desktop,
		Skin: resolveSkin("linux", version, nil, &desktop), Confidence: "exact",
	}
}

func (r *linuxReader) Appearance(skin string) Appearance {
	state := Appearance{Skin: skin, Accent: stockAccent(skin), Scheme: "light", Contrast: "normal", TextScale: 1}

	conn := r.connection()
	if conn == nil {
		return state
	}
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()
	read := func(namespace, key string) (interface{}, bool) { return readSetting(ctx, conn, namespace, key) }

	if v, ok := read(nsAppearance, "color-scheme"); ok {
		if n, ok := v.(uint32); ok && n == 1 { // 1 = prefer dark
			state.Scheme = "dark"
		}
	}
	if v, ok := read(nsAppearance, "accent-color"); ok {
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
	if v, ok := read(nsAppearance, "contrast"); ok {
		if n, ok := v.(uint32); ok && n == 1 { // 1 = higher contrast
			state.Contrast = "more"
		}
	}
	if v, ok := read(nsAppearance, "reduced-motion"); ok {
		if n, ok := v.(uint32); ok && n == 1 {
			state.ReduceMotion = true
		}
	}
	if v, ok := read(nsGnome, "text-scaling-factor"); ok {
		if f, ok := v.(float64); ok {
			state.TextScale = roundScale(f)
		}
	}
	return state
}

func (r *linuxReader) Capabilities() map[string]CapabilityLevel {
	// nativeMenus is partial: context menus are native, but a frameless window (custom title
	// bar) doesn't show the window menu bar, so desktop-ui draws the menubar in the page.
	return capabilities(map[string]CapabilityLevel{
		"systemAccent": Native, "colorScheme": Native, "highContrast": Native, "textScale": Partial,
		"windowFocus": Native, "customTitleBar": Native, "nativeMenus": Partial,
		"notifications": Partial, "fileDialogs": Native, "osVersion": Native, "linuxDesktop": Native,
	})
}

// connection returns the session bus, connecting again when there is none yet or it was
// lost (at most every reconnectEvery), so a bus that wasn't ready at startup is picked up.
func (r *linuxReader) connection() *dbus.Conn {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.conn != nil && r.conn.Connected() {
		return r.conn
	}
	if r.conn != nil {
		_ = r.conn.Close()
		r.conn = nil
	}
	if !r.lastAttempt.IsZero() && time.Since(r.lastAttempt) < reconnectEvery {
		return nil
	}
	r.lastAttempt = time.Now()
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil
	}
	r.conn = conn
	return conn
}

// readSetting asks the portal for one setting. ReadOne (portal v2) first, then the older Read.
func readSetting(ctx context.Context, conn *dbus.Conn, namespace, key string) (interface{}, bool) {
	obj := conn.Object(portalDest, portalPath)
	var v dbus.Variant
	if err := obj.CallWithContext(ctx, portalSetting+".ReadOne", 0, namespace, key).Store(&v); err == nil {
		return v.Value(), true
	}
	if ctx.Err() != nil {
		return nil, false
	}
	if err := obj.CallWithContext(ctx, portalSetting+".Read", 0, namespace, key).Store(&v); err == nil {
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
