package spartannative

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestToAccelerator(t *testing.T) {
	cases := map[string]string{
		"mod+shift+s": "cmdorctrl+shift+s",
		"Mod+N":       "cmdorctrl+n",
		"alt+f4":      "optionoralt+f4",
		"ctrl+delete": "ctrl+delete",
		"f2":          "f2",
		"enter":       "enter",
		"mod+,":       "cmdorctrl+,",
	}
	for in, want := range cases {
		got, ok := toAccelerator(in)
		if !ok || got != want {
			t.Errorf("toAccelerator(%q) = %q, %v; want %q", in, got, ok, want)
		}
	}
	for _, bad := range []string{"", "hyper+s", "mod+pageup", "mod+"} {
		if got, ok := toAccelerator(bad); ok {
			t.Errorf("toAccelerator(%q) = %q, want rejected", bad, got)
		}
	}
}

func TestResolveMenuDropsWhatTheOSCannotShow(t *testing.T) {
	sep := MenuItem{Type: "separator"}
	items := []MenuItem{
		{Role: "appMenu"},
		{Label: "File", Submenu: []MenuItem{
			sep,
			{ID: "new", Label: "New", Accelerator: "mod+n"},
			sep, sep,
			{Role: "hide"},
			{Role: "nonsense"},
			{Role: "quit"},
			sep,
		}},
		{Label: "Empty", Submenu: []MenuItem{{Role: "hideOthers"}}},
		{Role: "editMenu"},
	}

	labels := func(menu []MenuItem) []string {
		out := []string{}
		for _, i := range menu {
			switch {
			case i.Type == "separator":
				out = append(out, "—")
			case i.Role != "":
				out = append(out, "role:"+i.Role)
			default:
				out = append(out, i.Label)
			}
		}
		return out
	}

	win := resolveMenu(items, "windows")
	if got, want := labels(win), []string{"File", "role:editMenu"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("windows top level = %v, want %v", got, want)
	}
	if got, want := labels(win[0].Submenu), []string{"New", "—", "role:quit"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("windows File = %v, want %v", got, want)
	}

	mac := resolveMenu(items, "darwin")
	if got, want := labels(mac), []string{"role:appMenu", "File", "Empty", "role:editMenu"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("macOS top level = %v, want %v", got, want)
	}
	if got, want := labels(mac[1].Submenu), []string{"New", "—", "role:hide", "role:quit"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("macOS File = %v, want %v", got, want)
	}
}

func TestMenuItemJSONMatchesTheContract(t *testing.T) {
	var items []MenuItem
	src := `[{"label":"File","submenu":[{"id":"save","label":"Save","accelerator":"mod+s","enabled":false},{"type":"separator"},{"id":"wrap","label":"Wrap","type":"checkbox","checked":true}]}]`
	if err := json.Unmarshal([]byte(src), &items); err != nil {
		t.Fatal(err)
	}
	save := items[0].Submenu[0]
	if save.ID != "save" || save.Accelerator != "mod+s" || save.Enabled == nil || *save.Enabled {
		t.Fatalf("save = %+v", save)
	}
	if wrap := items[0].Submenu[2]; wrap.Type != "checkbox" || !wrap.Checked {
		t.Fatalf("wrap = %+v", wrap)
	}
}

func TestBuildMenuReportsChosenItems(t *testing.T) {
	disabled := false
	chosen := make(chan string, 1)
	menu := buildMenu(resolveMenu([]MenuItem{
		{ID: "new", Label: "New", Accelerator: "mod+n"},
		{Type: "separator"},
		{ID: "wrap", Label: "Wrap lines", Type: "checkbox", Checked: true},
		{ID: "old", Label: "Old", Enabled: &disabled},
		{Label: "More", Submenu: []MenuItem{{ID: "deep", Label: "Deep"}}},
	}, "linux"), func(id string) { chosen <- id })

	if menu.ItemAt(0).Label() != "New" || menu.ItemAt(0).GetAccelerator() == "" {
		t.Fatalf("New: label %q, accelerator %q", menu.ItemAt(0).Label(), menu.ItemAt(0).GetAccelerator())
	}
	if !menu.ItemAt(1).IsSeparator() {
		t.Fatal("item 1 is not a separator")
	}
	if wrap := menu.ItemAt(2); !wrap.IsCheckbox() || !wrap.Checked() {
		t.Fatal("Wrap lines is not a checked checkbox")
	}
	if menu.ItemAt(3).Enabled() {
		t.Fatal("Old is enabled")
	}
	more := menu.ItemAt(4)
	if !more.IsSubmenu() || more.GetSubmenu().ItemAt(0).Label() != "Deep" {
		t.Fatal("More has no Deep submenu")
	}
}
