//go:build windows

package spartannative

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// Windows: the registry keys the Settings app writes, and SystemParametersInfo.

const (
	keyPersonalize = `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`
	keyAccent      = `Software\Microsoft\Windows\CurrentVersion\Explorer\Accent`
	keyDWM         = `Software\Microsoft\Windows\DWM`
	keyAccess      = `Software\Microsoft\Accessibility`

	spiGetHighContrast       = 0x0042
	spiGetClientAreaAnimation = 0x1042
	hcfHighContrastOn        = 0x00000001
)

var procSystemParametersInfo = windows.NewLazySystemDLL("user32.dll").NewProc("SystemParametersInfoW")

type highContrast struct {
	cbSize            uint32
	dwFlags           uint32
	lpszDefaultScheme *uint16
}

type windowsReader struct{}

func newReader() reader { return windowsReader{} }

func (windowsReader) platform() PlatformInfo {
	v := windows.RtlGetVersion()
	version := fmt.Sprintf("%d.%d.%d", v.MajorVersion, v.MinorVersion, v.BuildNumber)
	build := int(v.BuildNumber)
	return PlatformInfo{
		OS: "windows", Version: &version, Build: &build,
		Skin: resolveSkin("windows", &version, &build, nil), Confidence: "exact",
	}
}

func (windowsReader) appearance(skin string) AppearanceState {
	scheme := "light"
	if v, ok := readDWORD(keyPersonalize, "AppsUseLightTheme"); ok && v == 0 {
		scheme = "dark"
	}
	transparency, ok := readDWORD(keyPersonalize, "EnableTransparency")
	reduceTransparency := ok && transparency == 0

	textScale := 1.0
	if v, ok := readDWORD(keyAccess, "TextScaleFactor"); ok && v >= 100 {
		textScale = roundScale(float64(v) / 100)
	}

	contrast := "normal"
	hc := highContrast{cbSize: uint32(unsafe.Sizeof(highContrast{}))}
	if r, _, _ := procSystemParametersInfo.Call(spiGetHighContrast, uintptr(hc.cbSize), uintptr(unsafe.Pointer(&hc)), 0); r != 0 && hc.dwFlags&hcfHighContrastOn != 0 {
		contrast = "forced" // a Windows contrast theme: system colors take over
	}

	var animations int32 = 1
	if r, _, _ := procSystemParametersInfo.Call(spiGetClientAreaAnimation, 0, uintptr(unsafe.Pointer(&animations)), 0); r == 0 {
		animations = 1
	}

	return AppearanceState{
		Skin:               skin,
		Accent:             windowsAccent(skin),
		Scheme:             scheme,
		Contrast:           contrast,
		ReduceTransparency: reduceTransparency,
		ReduceMotion:       animations == 0,
		TextScale:          textScale,
	}
}

func (windowsReader) capabilities() map[string]CapabilityLevel {
	return map[string]CapabilityLevel{
		"systemAccent": Native, "colorScheme": Native, "highContrast": Native, "reduceTransparency": Native,
		"textScale": Native, "windowFocus": Native, "windowMaterial": Native, "customTitleBar": Native,
		"nativeMenus": Partial, "notifications": Partial, "taskbarProgress": None, "fileDialogs": Native,
		"osVersion": Native, "linuxDesktop": None,
	}
}

// windowsAccent reads the accent the Settings app picked: AccentPalette holds eight
// RGBA swatches, the fourth (offset 12) is the base accent. DWM's AccentColor
// (0xAABBGGRR) is the fallback.
func windowsAccent(skin string) string {
	if k, err := registry.OpenKey(registry.CURRENT_USER, keyAccent, registry.QUERY_VALUE); err == nil {
		defer k.Close()
		if palette, _, err := k.GetBinaryValue("AccentPalette"); err == nil && len(palette) >= 16 {
			return fmt.Sprintf("#%02x%02x%02x", palette[12], palette[13], palette[14])
		}
	}
	if v, ok := readDWORD(keyDWM, "AccentColor"); ok {
		return fmt.Sprintf("#%02x%02x%02x", v&0xff, (v>>8)&0xff, (v>>16)&0xff)
	}
	return stockAccent(skin)
}

func readDWORD(path, name string) (uint32, bool) {
	k, err := registry.OpenKey(registry.CURRENT_USER, path, registry.QUERY_VALUE)
	if err != nil {
		return 0, false
	}
	defer k.Close()
	v, _, err := k.GetIntegerValue(name)
	return uint32(v), err == nil
}
