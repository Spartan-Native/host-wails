package appearance

import (
	"sync"
	"time"
)

// Watcher re-reads the appearance on a timer (and whenever Check is called) and reports
// it only when something changed since the last read.
type Watcher struct {
	read     func() Appearance
	onChange func(Appearance)
	interval time.Duration

	mu   sync.Mutex
	last *Appearance
	stop chan struct{}
}

// NewWatcher creates a watcher. read returns the current appearance; onChange receives
// it when it differs from the last one seen. An interval of zero or less turns polling
// off: only Check reports changes then.
func NewWatcher(interval time.Duration, read func() Appearance, onChange func(Appearance)) *Watcher {
	return &Watcher{read: read, onChange: onChange, interval: interval}
}

// Seen records a state the frontend already has, so it isn't reported again.
func (w *Watcher) Seen(state Appearance) {
	w.mu.Lock()
	w.last = &state
	w.mu.Unlock()
}

// Check re-reads the appearance and reports it if it changed.
func (w *Watcher) Check() {
	next := w.read()
	w.mu.Lock()
	changed := w.last == nil || *w.last != next
	if changed {
		w.last = &next
	}
	w.mu.Unlock()
	if changed && w.onChange != nil {
		w.onChange(next)
	}
}

// Start begins polling. Calling it again while running does nothing, and so does a
// watcher whose interval is zero or less.
func (w *Watcher) Start() {
	if w.interval <= 0 {
		return
	}
	w.mu.Lock()
	if w.stop != nil {
		w.mu.Unlock()
		return
	}
	w.stop = make(chan struct{})
	stop := w.stop
	w.mu.Unlock()

	go func() {
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				w.Check()
			}
		}
	}()
}

// Stop ends polling.
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.stop != nil {
		close(w.stop)
		w.stop = nil
	}
}
