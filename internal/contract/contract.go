// Package contract holds the data the Go service exchanges with the frontend. Every type
// mirrors the type of the same name in @spartan-native/theme and must keep producing the
// JSON it describes. The public package re-exports them as aliases.
package contract

// Events the service emits to the frontend.
const (
	// EventAppearance is emitted with an AppearanceState whenever a value changes.
	EventAppearance = "spartan-native:appearance"
	// EventMenuAction is emitted with an item's id when an item with an id is chosen,
	// in the application menu or a context menu.
	EventMenuAction = "spartan-native:menu-action"
)

// --- Appearance ---------------------------------------------------------------------

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

// --- Menus --------------------------------------------------------------------------

// MenuItem mirrors MenuItem in @spartan-native/theme.
type MenuItem struct {
	ID          string     `json:"id,omitempty"`
	Label       string     `json:"label,omitempty"`
	Type        string     `json:"type,omitempty"`        // normal | checkbox | radio | separator
	Accelerator string     `json:"accelerator,omitempty"` // portable notation: mod+shift+s
	Enabled     *bool      `json:"enabled,omitempty"`     // default true
	Checked     bool       `json:"checked,omitempty"`
	Role        string     `json:"role,omitempty"`
	Submenu     []MenuItem `json:"submenu,omitempty"`
}

// --- Dialogs ------------------------------------------------------------------------

// FileFilter mirrors FileFilter in @spartan-native/theme.
type FileFilter struct {
	Name       string   `json:"name"`
	Extensions []string `json:"extensions"` // without the dot; "*" matches every file
}

// OpenFileOptions mirrors OpenFileOptions in @spartan-native/theme.
type OpenFileOptions struct {
	Title       string       `json:"title,omitempty"`
	ButtonLabel string       `json:"buttonLabel,omitempty"`
	Directory   string       `json:"directory,omitempty"`
	Filters     []FileFilter `json:"filters,omitempty"`
	Multiple    bool         `json:"multiple,omitempty"`
}

// SaveFileOptions mirrors SaveFileOptions in @spartan-native/theme.
type SaveFileOptions struct {
	Title       string       `json:"title,omitempty"`
	ButtonLabel string       `json:"buttonLabel,omitempty"`
	Directory   string       `json:"directory,omitempty"`
	Filters     []FileFilter `json:"filters,omitempty"`
	FileName    string       `json:"fileName,omitempty"`
}

// PickFolderOptions mirrors PickFolderOptions in @spartan-native/theme.
type PickFolderOptions struct {
	Title       string `json:"title,omitempty"`
	ButtonLabel string `json:"buttonLabel,omitempty"`
	Directory   string `json:"directory,omitempty"`
}

// PickedFile mirrors PickedFile in @spartan-native/theme (native hosts always set Path).
type PickedFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// MessageButton is one button of an alert.
type MessageButton struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// MessageOptions mirrors MessageOptions in @spartan-native/theme.
type MessageOptions struct {
	Kind      string          `json:"kind,omitempty"` // info | question | warning | error
	Title     string          `json:"title,omitempty"`
	Message   string          `json:"message"`
	Buttons   []MessageButton `json:"buttons,omitempty"`
	DefaultID string          `json:"defaultId,omitempty"`
	CancelID  string          `json:"cancelId,omitempty"`
}
