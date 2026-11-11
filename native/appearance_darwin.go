//go:build darwin

package spartannative

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework AppKit -framework Foundation
#import <AppKit/AppKit.h>

typedef struct {
	double r, g, b;
	int hasAccent, dark, reduceTransparency, increaseContrast, reduceMotion;
	long major, minor, patch;
} snAppearance;

static snAppearance snRead(void) {
	snAppearance a = {0};
	@autoreleasepool {
		// The accent the user picked in System Settings → Appearance, in sRGB
		NSColor *accent = [[NSColor controlAccentColor] colorUsingColorSpace:[NSColorSpace sRGBColorSpace]];
		if (accent) {
			a.hasAccent = 1;
			a.r = accent.redComponent;
			a.g = accent.greenComponent;
			a.b = accent.blueComponent;
		}
		// Light / dark: the global setting, independent of this app's own appearance
		NSString *style = [[NSUserDefaults standardUserDefaults] stringForKey:@"AppleInterfaceStyle"];
		a.dark = style != nil && [style caseInsensitiveCompare:@"Dark"] == NSOrderedSame;

		NSWorkspace *ws = [NSWorkspace sharedWorkspace];
		a.reduceTransparency = ws.accessibilityDisplayShouldReduceTransparency;
		a.increaseContrast = ws.accessibilityDisplayShouldIncreaseContrast;
		a.reduceMotion = ws.accessibilityDisplayShouldReduceMotion;

		NSOperatingSystemVersion v = [[NSProcessInfo processInfo] operatingSystemVersion];
		a.major = v.majorVersion;
		a.minor = v.minorVersion;
		a.patch = v.patchVersion;
	}
	return a;
}
*/
import "C"

import "fmt"

// macOS: AppKit through cgo. There is no global text size on macOS (apps offer
// their own), so textScale stays 1 and the capability is reported as none.

type darwinReader struct{}

func newReader() reader { return darwinReader{} }

func (darwinReader) platform() PlatformInfo {
	a := C.snRead()
	version := fmt.Sprintf("%d.%d.%d", int(a.major), int(a.minor), int(a.patch))
	return PlatformInfo{
		OS: "macos", Version: &version,
		Skin: resolveSkin("macos", &version, nil, nil), Confidence: "exact",
	}
}

func (darwinReader) appearance(skin string) AppearanceState {
	a := C.snRead()
	accent := stockAccent(skin)
	if a.hasAccent != 0 {
		accent = rgbHex(float64(a.r), float64(a.g), float64(a.b))
	}
	scheme, contrast := "light", "normal"
	if a.dark != 0 {
		scheme = "dark"
	}
	if a.increaseContrast != 0 {
		contrast = "more"
	}
	return AppearanceState{
		Skin:               skin,
		Accent:             accent,
		Scheme:             scheme,
		Contrast:           contrast,
		ReduceTransparency: a.reduceTransparency != 0,
		ReduceMotion:       a.reduceMotion != 0,
		TextScale:          1,
	}
}

func (darwinReader) capabilities() map[string]CapabilityLevel {
	return map[string]CapabilityLevel{
		"systemAccent": Native, "colorScheme": Native, "highContrast": Native, "reduceTransparency": Native,
		"textScale": None, "windowFocus": Native, "windowMaterial": Native, "customTitleBar": Native,
		"nativeMenus": Partial, "notifications": Partial, "taskbarProgress": None, "fileDialogs": Native,
		"osVersion": Native, "linuxDesktop": None,
	}
}
