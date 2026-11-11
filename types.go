package spartannative

import "github.com/spartan-native/host-wails/internal/contract"

// Events the service emits to the frontend.
const (
	// EventAppearance is emitted with an AppearanceState whenever a value changes.
	EventAppearance = contract.EventAppearance
	// EventMenuAction is emitted with an item's id when an item with an id is chosen,
	// in the application menu or a context menu.
	EventMenuAction = contract.EventMenuAction
)

// The data the service exchanges with the frontend. Each type mirrors the type of the same
// name in @spartan-native/theme; see internal/contract for the fields.
type (
	PlatformInfo      = contract.PlatformInfo
	AppearanceState   = contract.AppearanceState
	CapabilityLevel   = contract.CapabilityLevel
	MenuItem          = contract.MenuItem
	FileFilter        = contract.FileFilter
	OpenFileOptions   = contract.OpenFileOptions
	SaveFileOptions   = contract.SaveFileOptions
	PickFolderOptions = contract.PickFolderOptions
	PickedFile        = contract.PickedFile
	MessageButton     = contract.MessageButton
	MessageOptions    = contract.MessageOptions
)

// Capability levels.
const (
	Native  = contract.Native
	Partial = contract.Partial
	None    = contract.None
)
