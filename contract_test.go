package spartannative

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/spartan-native/host-wails/internal/contract"
)

// The names and JSON in testdata/contract are shared with the npm package's spec, so both
// halves are checked against the same files.

type contractNames struct {
	Service string `json:"service"`
	Events  struct {
		Appearance string `json:"appearance"`
		MenuAction string `json:"menuAction"`
	} `json:"events"`
	Capabilities []string `json:"capabilities"`
}

func readFixture(t *testing.T, name string, v any) {
	t.Helper()
	raw, err := os.ReadFile("testdata/contract/" + name)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, v); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
}

func TestNamesMatchTheFixture(t *testing.T) {
	var names contractNames
	readFixture(t, "names.json", &names)

	// Wails binds methods as <package path>.<type>.<method>.
	if got := reflect.TypeOf(Service{}).PkgPath() + ".Service"; got != names.Service {
		t.Errorf("bound service name %q, fixture says %q", got, names.Service)
	}
	if EventAppearance != names.Events.Appearance || EventMenuAction != names.Events.MenuAction {
		t.Errorf("events %q, %q; fixture says %+v", EventAppearance, EventMenuAction, names.Events)
	}
	if !slices.Equal(contract.Capabilities, names.Capabilities) {
		t.Errorf("contract.Capabilities = %v\nfixture says %v", contract.Capabilities, names.Capabilities)
	}
	// This OS's reader reports exactly those keys.
	got := New().Capabilities()
	keys := make([]string, 0, len(got))
	for k := range got {
		keys = append(keys, k)
	}
	want := slices.Clone(names.Capabilities)
	slices.Sort(keys)
	slices.Sort(want)
	if !slices.Equal(keys, want) {
		t.Errorf("Capabilities() keys = %v, want %v", keys, want)
	}
}

// TestValuesRoundTrip decodes each fixture value into its Go type, rejecting fields Go
// doesn't have, and checks that encoding it again gives the same JSON.
func TestValuesRoundTrip(t *testing.T) {
	var values map[string]json.RawMessage
	readFixture(t, "values.json", &values)

	types := map[string]func() any{
		"platformInfo":      func() any { return new(PlatformInfo) },
		"appearanceState":   func() any { return new(AppearanceState) },
		"menu":              func() any { return new([]MenuItem) },
		"openFileOptions":   func() any { return new(OpenFileOptions) },
		"saveFileOptions":   func() any { return new(SaveFileOptions) },
		"pickFolderOptions": func() any { return new(PickFolderOptions) },
		"pickedFile":        func() any { return new(PickedFile) },
		"messageOptions":    func() any { return new(MessageOptions) },
	}
	for name := range values {
		if types[name] == nil {
			t.Errorf("values.json has %q, which no Go type is checked against", name)
		}
	}
	for name, newValue := range types {
		raw, ok := values[name]
		if !ok {
			t.Errorf("values.json has no %q", name)
			continue
		}
		v := newValue()
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		if err := dec.Decode(v); err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		back, _ := json.Marshal(v)
		if !jsonEqual(t, raw, back) {
			t.Errorf("%s does not round-trip:\n fixture %s\n     got %s", name, compact(raw), back)
		}
	}
}

func jsonEqual(t *testing.T, a, b []byte) bool {
	t.Helper()
	var x, y any
	if err := json.Unmarshal(a, &x); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &y); err != nil {
		t.Fatal(err)
	}
	return reflect.DeepEqual(x, y)
}

func compact(raw []byte) string {
	var b bytes.Buffer
	_ = json.Compact(&b, raw)
	return b.String()
}
