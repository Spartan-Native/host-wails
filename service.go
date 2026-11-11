package spartannative

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Service reads the OS appearance for the frontend and reports changes.
type Service struct {
	reader       reader
	pollInterval time.Duration
	skinOverride string
	emit         func(state AppearanceState)
	emitAction   func(id string)

	mu       sync.Mutex
	platform *PlatformInfo
	last     *AppearanceState
	stop     chan struct{}
	offTheme func()
}

// Option configures the service.
type Option func(*Service)

// WithPollInterval sets how often the OS is re-read for changes that have no
// notification (accent, transparency, text size). Default: 1 second.
func WithPollInterval(d time.Duration) Option { return func(s *Service) { s.pollInterval = d } }

// WithSkin forces a skin instead of detecting it (e.g. from a user setting).
func WithSkin(skin string) Option { return func(s *Service) { s.skinOverride = skin } }

// debug logs every call and change when SPARTAN_NATIVE_DEBUG is set.
var debug = os.Getenv("SPARTAN_NATIVE_DEBUG") != ""

func debugf(format string, args ...any) {
	if debug {
		log.Printf("[spartan-native] "+format, args...)
	}
}

// New creates the service with the reader for the current OS.
func New(opts ...Option) *Service {
	s := &Service{reader: newReader(), pollInterval: time.Second}
	for _, o := range opts {
		o(s)
	}
	return s
}

// --- Bound methods (called from @spartan-native/host-wails) -------------------

// Platform reports the OS, its version and the matching skin.
func (s *Service) Platform() PlatformInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.platform == nil {
		p := s.reader.platform()
		if s.skinOverride != "" {
			p.Skin = s.skinOverride
		}
		s.platform = &p
		debugf("Platform() → %s %v skin=%s", p.OS, deref(p.Version), p.Skin)
	}
	return *s.platform
}

// Appearance reports the current OS appearance.
func (s *Service) Appearance() AppearanceState {
	state := s.reader.appearance(s.Platform().Skin)
	s.mu.Lock()
	s.last = &state
	s.mu.Unlock()
	debugf("Appearance() → %+v", state)
	return state
}

// Capabilities reports what this OS lets the adapter read natively.
func (s *Service) Capabilities() map[string]CapabilityLevel {
	return s.reader.capabilities()
}

// --- Wails lifecycle -------------------------------------------------------------

// ServiceStartup starts watching for changes. Called by Wails.
func (s *Service) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	app := application.Get()
	if s.emit == nil && app != nil {
		s.emit = func(state AppearanceState) { app.Event.Emit(EventAppearance, state) }
	}
	if s.emitAction == nil && app != nil {
		s.emitAction = func(id string) { app.Event.Emit(EventMenuAction, id) }
	}
	if app != nil {
		// The OS tells Wails about light/dark changes; check everything then.
		s.offTheme = app.Event.OnApplicationEvent(events.Common.ThemeChanged, func(*application.ApplicationEvent) { s.check() })
	}
	s.start()
	return nil
}

// ServiceShutdown stops watching. Called by Wails.
func (s *Service) ServiceShutdown() error {
	if s.offTheme != nil {
		s.offTheme()
	}
	s.stopWatching()
	return nil
}

func (s *Service) start() {
	s.mu.Lock()
	if s.stop != nil {
		s.mu.Unlock()
		return
	}
	s.stop = make(chan struct{})
	stop := s.stop
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(s.pollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				s.check()
			}
		}
	}()
}

func (s *Service) stopWatching() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stop != nil {
		close(s.stop)
		s.stop = nil
	}
}

// check re-reads the OS and emits when anything changed.
func (s *Service) check() {
	next := s.reader.appearance(s.Platform().Skin)
	s.mu.Lock()
	changed := s.last == nil || *s.last != next
	if changed {
		s.last = &next
	}
	emit := s.emit
	s.mu.Unlock()
	if changed && emit != nil {
		debugf("changed → emit %s %+v", EventAppearance, next)
		emit(next)
	}
}
