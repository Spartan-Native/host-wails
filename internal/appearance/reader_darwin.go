//go:build darwin

package appearance

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework AppKit -framework Foundation
#import <AppKit/AppKit.h>

typedef struct {
	double r, g, b;
	int hasAccent, dark, reduceTransparency, increaseContrast, reduceMotion;
	long major, minor, patch;
} snAppearance;

static void snReadInto(snAppearance *out) {
	snAppearance a = {0};
	@autoreleasepool {
		// The accent the user picked in System Settings → Appearance, in sRGB. Resolved in the
		// light (Aqua) appearance, so the value doesn't depend on which thread reads it (the
		// main thread would otherwise give the dark variant in dark mode).
		__block NSColor *accent = nil;
		[[NSAppearance appearanceNamed:NSAppearanceNameAqua] performAsCurrentDrawingAppearance:^{
			accent = [[NSColor controlAccentColor] colorUsingColorSpace:[NSColorSpace sRGBColorSpace]];
		}];
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
	*out = a;
}

// AppKit isn't thread-safe, and the readers run on goroutines. While the app runs, read on
// the main thread; without a running app (tests, before startup) there is no main run loop
// to wait for, so read here.
static snAppearance snRead(void) {
	__block snAppearance a = {0};
	if ([NSThread isMainThread] || NSApp == nil || ![NSApp isRunning]) {
		snReadInto(&a);
	} else {
		dispatch_sync(dispatch_get_main_queue(), ^{ snReadInto(&a); });
	}
	return a;
}
*/
import "C"

import "fmt"

// macOS: AppKit through cgo. There is no global text size on macOS (apps offer
// their own), so textScale stays 1 and the capability is reported as none.

type darwinReader struct{}

func New() Reader { return darwinReader{} }

func (darwinReader) Platform() Platform {
	a := C.snRead()
	version := fmt.Sprintf("%d.%d.%d", int(a.major), int(a.minor), int(a.patch))
	return Platform{
		OS: "macos", Version: &version,
		Skin: resolveSkin("macos", &version, nil, nil), Confidence: "exact",
	}
}

func (darwinReader) Appearance(skin string) Appearance {
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
	return Appearance{
		Skin:               skin,
		Accent:             accent,
		Scheme:             scheme,
		Contrast:           contrast,
		ReduceTransparency: a.reduceTransparency != 0,
		ReduceMotion:       a.reduceMotion != 0,
		TextScale:          1,
	}
}

// Capabilities lists what macOS supports; the rest is None.
func (darwinReader) Capabilities() map[string]CapabilityLevel {
	return capabilities(map[string]CapabilityLevel{
		"systemAccent": Native, "colorScheme": Native, "highContrast": Native,
		"reduceTransparency": Native, "windowFocus": Native, "windowMaterial": Native,
		"customTitleBar": Native, "nativeMenus": Native, "notifications": Partial, "fileDialogs": Native,
		"osVersion": Native,
	})
}
