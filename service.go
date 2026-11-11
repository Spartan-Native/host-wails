package spartannative

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/spartan-native/host-wails/internal/appearance"
	"github.com/spartan-native/host-wails/internal/debug"
	"github.com/spartan-native/host-wails/internal/dialogs"
	"github.com/spartan-native/host-wails/internal/menus"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Service is the Wails service the frontend adapter talks to. Its exported methods are
// bound by Wails as github.com/spartan-native/host-wails.Service.<Method>, the names
// @spartan-native/host-wails calls, so the type must stay in this package.
//
// The work happens in internal/appearance, internal/menus and internal/dialogs; this
// type wires them to Wails.
type Service struct {
	reader       appearance.Reader
	pollInterval time.Duration
	skinOverride string
	emit         func(state AppearanceState)
	emitAction   func(id string)

	mu       sync.Mutex
	platform *PlatformInfo
	watcher  *appearance.Watcher
	offTheme func()
}

// Option configures the service.
type Option func(*Service)

// WithPollInterval sets how often the OS is re-read for changes that have no
// notification (accent, transparency, text size). Default: 1 second. Zero or less turns
// polling off; changes are then picked up only when Wails reports a theme change.
func WithPollInterval(d time.Duration) Option { return func(s *Service) { s.pollInterval = d } }

// WithSkin forces a skin instead of detecting it (e.g. from a user setting).
func WithSkin(skin string) Option { return func(s *Service) { s.skinOverride = skin } }

// New creates the service with the reader for the current OS.
func New(opts ...Option) *Service {
	s := &Service{reader: appearance.New(), pollInterval: time.Second}
	for _, o := range opts {
		o(s)
	}
	s.watcher = appearance.NewWatcher(s.pollInterval, s.read, func(state AppearanceState) {
		debug.Logf("changed → emit %s %+v", EventAppearance, state)
		if s.emit != nil {
			s.emit(state)
		}
	})
	return s
}

var errNoApp = errors.New("spartan-native: no running Wails application")

// app returns the running Wails application.
func app() (*application.App, error) {
	if a := application.Get(); a != nil {
		return a, nil
	}
	return nil, errNoApp
}

// --- Wails lifecycle -------------------------------------------------------------

// ServiceStartup starts watching for changes. Called by Wails.
func (s *Service) ServiceStartup(context.Context, application.ServiceOptions) error {
	if a := application.Get(); a != nil {
		if s.emit == nil {
			s.emit = func(state AppearanceState) { a.Event.Emit(EventAppearance, state) }
		}
		if s.emitAction == nil {
			s.emitAction = func(id string) { a.Event.Emit(EventMenuAction, id) }
		}
		// The OS tells Wails about light/dark changes; check everything then. On a goroutine:
		// the handler may run on the main thread, which a running check can be waiting for.
		s.offTheme = a.Event.OnApplicationEvent(events.Common.ThemeChanged, func(*application.ApplicationEvent) { go s.watcher.Check() })
	}
	s.watcher.Start()
	return nil
}

// ServiceShutdown stops watching. Called by Wails.
func (s *Service) ServiceShutdown() error {
	if s.offTheme != nil {
		s.offTheme()
	}
	s.watcher.Stop()
	return nil
}

// --- Bound methods: appearance -----------------------------------------------------

// Platform reports the OS, its version and the matching skin.
func (s *Service) Platform() PlatformInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.platform == nil {
		p := s.reader.Platform()
		if s.skinOverride != "" {
			p.Skin = s.skinOverride
		}
		s.platform = &p
		debug.Logf("Platform() → %s %v skin=%s", p.OS, deref(p.Version), p.Skin)
	}
	return *s.platform
}

// Appearance reports the current OS appearance.
func (s *Service) Appearance() AppearanceState {
	state := s.read()
	s.watcher.Seen(state)
	debug.Logf("Appearance() → %+v", state)
	return state
}

// Capabilities reports what this OS lets the adapter read natively.
func (s *Service) Capabilities() map[string]CapabilityLevel {
	return s.reader.Capabilities()
}

func (s *Service) read() AppearanceState { return s.reader.Appearance(s.Platform().Skin) }

// --- Bound methods: menus ----------------------------------------------------------

// SetApplicationMenu sets the global menu bar on macOS. On Windows and Linux, Wails 3.1 gives the
// application menu to windows created after the call. An empty menu restores Wails' default.
func (s *Service) SetApplicationMenu(items []MenuItem) error {
	a, err := app()
	if err != nil {
		return err
	}
	menus.SetApplication(a, items, s.menuAction)
	return nil
}

// ShowContextMenu opens a native context menu at x, y (CSS pixels in the current window).
func (s *Service) ShowContextMenu(items []MenuItem, x, y int) error {
	a, err := app()
	if err != nil {
		return err
	}
	return menus.ShowContext(a, items, x, y, s.menuAction)
}

// menuAction reports a chosen item to the frontend.
func (s *Service) menuAction(id string) {
	debug.Logf("menu action %q", id)
	if s.emitAction != nil {
		s.emitAction(id)
	}
}

// --- Bound methods: dialogs --------------------------------------------------------

// OpenFile shows the OS's open dialog. A cancelled dialog returns no files.
func (s *Service) OpenFile(o OpenFileOptions) ([]PickedFile, error) {
	a, err := app()
	if err != nil {
		return nil, err
	}
	return dialogs.OpenFile(a, o)
}

// SaveFile shows the OS's save dialog. A cancelled dialog returns nil.
func (s *Service) SaveFile(o SaveFileOptions) (*PickedFile, error) {
	a, err := app()
	if err != nil {
		return nil, err
	}
	return dialogs.SaveFile(a, o)
}

// PickFolder shows the OS's folder dialog. A cancelled dialog returns nil.
func (s *Service) PickFolder(o PickFolderOptions) (*PickedFile, error) {
	a, err := app()
	if err != nil {
		return nil, err
	}
	return dialogs.PickFolder(a, o)
}

// Message shows the OS's alert and returns the id of the pressed button. If the alert is
// closed without one (GTK lets the user do that), the call ends when the frontend cancels it.
func (s *Service) Message(ctx context.Context, o MessageOptions) (*string, error) {
	a, err := app()
	if err != nil {
		return nil, err
	}
	return dialogs.Message(ctx, a, o)
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
