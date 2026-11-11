// Package menus builds native Wails menus from the contract's MenuItem data.
package menus

import (
	"errors"
	"runtime"
	"strconv"
	"strings"

	"github.com/spartan-native/host-wails/internal/contract"
	"github.com/spartan-native/host-wails/internal/debug"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// MenuItem is the contract's menu item.
type MenuItem = contract.MenuItem

// contextMenuName is the one context menu the package registers; each ShowContext replaces it.
const contextMenuName = "spartan-native:context-menu"

// roles maps the contract's MenuRole names to Wails roles.
var roles = map[string]application.Role{
	"appMenu":          application.AppMenu,
	"editMenu":         application.EditMenu,
	"windowMenu":       application.WindowMenu,
	"about":            application.About,
	"hide":             application.Hide,
	"hideOthers":       application.HideOthers,
	"showAll":          application.ShowAll,
	"quit":             application.Quit,
	"undo":             application.Undo,
	"redo":             application.Redo,
	"cut":              application.Cut,
	"copy":             application.Copy,
	"paste":            application.Paste,
	"selectAll":        application.SelectAll,
	"minimize":         application.Minimise,
	"zoom":             application.Zoom,
	"close":            application.CloseWindow,
	"toggleFullscreen": application.ToggleFullscreen,
}

// macOnlyRoles exist only in the macOS app menu.
var macOnlyRoles = map[string]bool{"appMenu": true, "hide": true, "hideOthers": true, "showAll": true}

// modifiers maps the portable notation to Wails' accelerator modifiers.
var modifiers = map[string]string{"mod": "cmdorctrl", "ctrl": "ctrl", "alt": "optionoralt", "shift": "shift", "meta": "super"}

// namedKeys maps the portable notation's key names to the names Wails knows. Most are the
// same; Wails spells the paging keys with a space.
var namedKeys = map[string]string{
	"backspace": "backspace", "tab": "tab", "enter": "enter", "escape": "escape", "space": "space",
	"delete": "delete", "forwarddelete": "forward delete", "up": "up", "down": "down", "left": "left",
	"right": "right", "home": "home", "end": "end", "pageup": "page up", "pagedown": "page down",
	"plus": "plus", "numlock": "numlock",
}

// functionKey reports whether k is f1 … f35, the function keys Wails knows.
func functionKey(k string) bool {
	if len(k) < 2 || k[0] != 'f' || k[1] == '0' {
		return false
	}
	n, err := strconv.Atoi(k[1:])
	return err == nil && n >= 1 && n <= 35
}

// toAccelerator turns "mod+shift+s" into Wails' "cmdorctrl+shift+s". ok is false for
// shortcuts Wails can't take (an unknown key or modifier); the item then has none.
func toAccelerator(shortcut string) (accelerator string, ok bool) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(shortcut)), "+")
	if len(parts) == 0 || parts[0] == "" {
		return "", false
	}
	out := make([]string, 0, len(parts))
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if i < len(parts)-1 {
			m, known := modifiers[p]
			if !known {
				return "", false
			}
			out = append(out, m)
			continue
		}
		switch name, named := namedKeys[p]; {
		case named:
			p = name
		case len([]rune(p)) == 1, functionKey(p):
		default:
			return "", false
		}
		out = append(out, p)
	}
	return strings.Join(out, "+"), true
}

// resolveMenu drops what the OS can't show: roles it doesn't have, unknown roles, submenus
// left empty, and separators at the edges or next to each other. Pure, so it is tested
// without a running Wails application.
func resolveMenu(items []MenuItem, goos string) []MenuItem {
	out := make([]MenuItem, 0, len(items))
	for _, item := range items {
		if item.Type == "separator" {
			if len(out) == 0 || out[len(out)-1].Type == "separator" {
				continue
			}
			out = append(out, MenuItem{Type: "separator"})
			continue
		}
		if item.Role != "" {
			if _, known := roles[item.Role]; !known || (macOnlyRoles[item.Role] && goos != "darwin") {
				continue
			}
		}
		if item.Submenu != nil {
			item.Submenu = resolveMenu(item.Submenu, goos)
			if len(item.Submenu) == 0 {
				continue
			}
		}
		out = append(out, item)
	}
	for len(out) > 0 && out[len(out)-1].Type == "separator" {
		out = out[:len(out)-1]
	}
	return out
}

// buildMenu turns resolved items into a Wails menu. onAction gets the id of a chosen item.
func buildMenu(items []MenuItem, onAction func(id string)) *application.Menu {
	built := make([]*application.MenuItem, 0, len(items))
	for _, item := range items {
		if mi := buildItem(item, onAction); mi != nil {
			built = append(built, mi)
		}
	}
	if len(built) == 0 {
		return application.NewMenu()
	}
	return application.NewMenuFromItems(built[0], built[1:]...)
}

func buildItem(item MenuItem, onAction func(id string)) *application.MenuItem {
	var mi *application.MenuItem
	switch {
	case item.Type == "separator":
		return application.NewMenuItemSeparator()
	case item.Role != "":
		mi = application.NewRole(roles[item.Role])
		if mi == nil {
			return nil
		}
		if item.Label != "" {
			mi.SetLabel(item.Label)
		}
	case item.Submenu != nil:
		mi = application.NewSubmenu(item.Label, buildMenu(item.Submenu, onAction))
	case item.Type == "checkbox":
		mi = application.NewMenuItemCheckbox(item.Label, item.Checked)
	case item.Type == "radio":
		mi = application.NewMenuItemRadio(item.Label, item.Checked)
	default:
		mi = application.NewMenuItem(item.Label)
	}
	if item.Enabled != nil {
		mi.SetEnabled(*item.Enabled)
	}
	if item.Accelerator != "" && item.Submenu == nil {
		if acc, ok := toAccelerator(item.Accelerator); ok {
			mi.SetAccelerator(acc)
		} else {
			debug.Logf("menu item %q: accelerator %q not supported, dropped", item.Label, item.Accelerator)
		}
	}
	if item.ID != "" && item.Role == "" && item.Submenu == nil {
		id := item.ID
		mi.OnClick(func(*application.Context) { onAction(id) })
	}
	return mi
}

// SetApplication sets the global menu bar on macOS. On Windows and Linux, Wails 3.1 gives the
// application menu to windows created after the call. An empty menu restores Wails' default.
// onAction gets the id of a chosen item.
func SetApplication(app *application.App, items []MenuItem, onAction func(id string)) {
	resolved := resolveMenu(items, runtime.GOOS)
	debug.Logf("SetApplicationMenu(%d items)", len(resolved))
	application.InvokeSync(func() {
		if len(resolved) == 0 {
			app.Menu.SetApplicationMenu(nil) // Wails' default menu for the OS
			return
		}
		app.Menu.SetApplicationMenu(buildMenu(resolved, onAction))
	})
}

// ShowContext opens a native context menu at x, y (CSS pixels in the current window).
// onAction gets the id of a chosen item.
func ShowContext(app *application.App, items []MenuItem, x, y int, onAction func(id string)) error {
	window := app.Window.Current()
	if window == nil {
		if all := app.Window.GetAll(); len(all) > 0 {
			window = all[0]
		}
	}
	if window == nil {
		return errors.New("spartan-native: no window for the context menu")
	}
	resolved := resolveMenu(items, runtime.GOOS)
	debug.Logf("ShowContextMenu(%d items) at %d,%d", len(resolved), x, y)
	application.InvokeSync(func() {
		menu := application.NewContextMenu(contextMenuName)
		menu.Menu = buildMenu(resolved, onAction)
		menu.Update()
	})
	window.OpenContextMenu(&application.ContextMenuData{Id: contextMenuName, X: x, Y: y})
	return nil
}
