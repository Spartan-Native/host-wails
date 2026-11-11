package spartannative

import (
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"
)

type fakeReader struct {
	mu    sync.Mutex
	state AppearanceState
}

func (f *fakeReader) platform() PlatformInfo {
	v := "26.1.0"
	return PlatformInfo{OS: "macos", Version: &v, Skin: "mac", Confidence: "exact"}
}
func (f *fakeReader) appearance(skin string) AppearanceState {
	f.mu.Lock()
	defer f.mu.Unlock()
	s := f.state
	s.Skin = skin
	return s
}
func (f *fakeReader) capabilities() map[string]CapabilityLevel { return map[string]CapabilityLevel{"systemAccent": Native} }
func (f *fakeReader) set(fn func(*AppearanceState)) {
	f.mu.Lock()
	fn(&f.state)
	f.mu.Unlock()
}

func newTestService(r *fakeReader, emitted chan AppearanceState, opts ...Option) *Service {
	s := New(append(opts, WithPollInterval(10*time.Millisecond))...)
	s.reader = r
	s.emit = func(a AppearanceState) { emitted <- a }
	return s
}

func TestEmitsOnlyWhenSomethingChanges(t *testing.T) {
	r := &fakeReader{state: AppearanceState{Accent: "#007aff", Scheme: "light", Contrast: "normal", TextScale: 1}}
	emitted := make(chan AppearanceState, 8)
	s := newTestService(r, emitted)
	s.Appearance() // the frontend's first read
	s.start()
	defer s.stopWatching()

	select {
	case a := <-emitted:
		t.Fatalf("emitted without a change: %+v", a)
	case <-time.After(60 * time.Millisecond):
	}

	r.set(func(a *AppearanceState) { a.Scheme = "dark"; a.Accent = "#e66100" })
	select {
	case a := <-emitted:
		if a.Scheme != "dark" || a.Accent != "#e66100" || a.Skin != "mac" {
			t.Fatalf("unexpected state %+v", a)
		}
	case <-time.After(time.Second):
		t.Fatal("no event after the appearance changed")
	}
}

func TestSkinOverrideAndJSONShape(t *testing.T) {
	s := New(WithSkin("gtk"))
	s.reader = &fakeReader{state: AppearanceState{Accent: "#3584e4", Scheme: "light", Contrast: "normal", TextScale: 1}}
	if got := s.Platform().Skin; got != "gtk" {
		t.Fatalf("skin override ignored: %s", got)
	}
	raw, _ := json.Marshal(s.Appearance())
	want := `{"skin":"gtk","accent":"#3584e4","scheme":"light","contrast":"normal","reduceTransparency":false,"reduceMotion":false,"textScale":1}`
	if string(raw) != want {
		t.Fatalf("JSON shape changed:\n got %s\nwant %s", raw, want)
	}
}

func TestResolveSkin(t *testing.T) {
	v15, v26 := "15.6.1", "26.1.0"
	b10, b11 := 19045, 22631
	kde, gnome := "kde", "gnome"
	cases := []struct {
		os      string
		version *string
		build   *int
		desktop *string
		want    string
	}{
		{"windows", nil, &b11, nil, "win"},
		{"windows", nil, &b10, nil, "w10"},
		{"macos", &v26, nil, nil, "mac"},
		{"macos", &v15, nil, nil, "macc"},
		{"linux", nil, nil, &gnome, "gtk"},
		{"linux", nil, nil, &kde, "qt"},
		{"plan9", nil, nil, nil, "web"},
	}
	for _, c := range cases {
		if got := resolveSkin(c.os, c.version, c.build, c.desktop); got != c.want {
			t.Errorf("resolveSkin(%s) = %s, want %s", c.os, got, c.want)
		}
	}
}

func TestHelpers(t *testing.T) {
	if got := rgbHex(0, 0.478, 1); got != "#007aff" {
		t.Errorf("rgbHex = %s", got)
	}
	for in, want := range map[string]string{"ubuntu:GNOME": "gnome", "KDE": "kde", "X-Cinnamon": "cinnamon", "XFCE": "xfce", "sway": "other"} {
		if got := linuxDesktop(in); got != want {
			t.Errorf("linuxDesktop(%q) = %s, want %s", in, got, want)
		}
	}
	if roundScale(1.249) != 1.25 || roundScale(0) != 1 {
		t.Error("roundScale")
	}
}

// TestLiveRead prints what this machine reports. Run with SN_LIVE=1 go test -run TestLiveRead -v
func TestLiveRead(t *testing.T) {
	if os.Getenv("SN_LIVE") == "" {
		t.Skip("set SN_LIVE=1 to read this machine's real appearance")
	}
	s := New()
	p, _ := json.MarshalIndent(s.Platform(), "", "  ")
	a, _ := json.MarshalIndent(s.Appearance(), "", "  ")
	t.Logf("platform: %s", p)
	t.Logf("appearance: %s", a)
}
