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

func (f *fakeReader) Platform() PlatformInfo {
	v := "26.1.0"
	return PlatformInfo{OS: "macos", Version: &v, Skin: "mac", Confidence: "exact"}
}
func (f *fakeReader) Appearance(skin string) AppearanceState {
	f.mu.Lock()
	defer f.mu.Unlock()
	s := f.state
	s.Skin = skin
	return s
}
func (f *fakeReader) Capabilities() map[string]CapabilityLevel {
	return map[string]CapabilityLevel{"systemAccent": Native}
}
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
	s.watcher.Start()
	defer s.watcher.Stop()

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

func TestZeroPollIntervalTurnsPollingOff(t *testing.T) {
	r := &fakeReader{state: AppearanceState{Accent: "#007aff", Scheme: "light", Contrast: "normal", TextScale: 1}}
	emitted := make(chan AppearanceState, 8)
	s := New(WithPollInterval(0)) // must not panic
	s.reader = r
	s.emit = func(a AppearanceState) { emitted <- a }
	s.Appearance()
	s.watcher.Start() // no ticker
	defer s.watcher.Stop()

	r.set(func(a *AppearanceState) { a.Scheme = "dark" })
	select {
	case a := <-emitted:
		t.Fatalf("polled with polling off: %+v", a)
	case <-time.After(50 * time.Millisecond):
	}
	s.watcher.Check() // what a ThemeChanged event does
	if a := <-emitted; a.Scheme != "dark" {
		t.Fatalf("Check did not report the change: %+v", a)
	}
}
