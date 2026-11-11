package appearance

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWatcherStopsAndRestarts(t *testing.T) {
	var reads atomic.Int32
	w := NewWatcher(5*time.Millisecond, func() Appearance { reads.Add(1); return Appearance{} }, nil)

	w.Start()
	w.Start() // a second Start while running does nothing
	time.Sleep(30 * time.Millisecond)
	w.Stop()
	w.Stop() // a second Stop does nothing
	stopped := reads.Load()
	if stopped == 0 {
		t.Fatal("never polled")
	}
	time.Sleep(30 * time.Millisecond)
	if n := reads.Load(); n > stopped+1 { // one tick may already be running
		t.Fatalf("still polling after Stop: %d reads, was %d", n, stopped)
	}

	w.Start()
	defer w.Stop()
	time.Sleep(30 * time.Millisecond)
	if reads.Load() <= stopped+1 {
		t.Fatal("did not poll after a restart")
	}
}

func TestWatcherWithoutIntervalDoesNotPoll(t *testing.T) {
	var reads atomic.Int32
	for _, interval := range []time.Duration{0, -time.Second} {
		w := NewWatcher(interval, func() Appearance { reads.Add(1); return Appearance{} }, nil)
		w.Start()
		time.Sleep(20 * time.Millisecond)
		w.Stop()
	}
	if n := reads.Load(); n != 0 {
		t.Fatalf("polled %d times with polling off", n)
	}
}

// TestWatcherReportsTheLatestState runs overlapping checks. The first read is slow and
// returns the older state; it must not be reported after the newer one.
func TestWatcherReportsTheLatestState(t *testing.T) {
	var n atomic.Int32
	slowStarted := make(chan struct{})
	read := func() Appearance {
		if n.Add(1) == 1 {
			close(slowStarted)
			time.Sleep(20 * time.Millisecond)
			return Appearance{Scheme: "light"}
		}
		return Appearance{Scheme: "dark"}
	}
	var mu sync.Mutex
	var reported []string
	w := NewWatcher(0, read, func(a Appearance) {
		mu.Lock()
		reported = append(reported, a.Scheme)
		mu.Unlock()
	})

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); w.Check() }()
	<-slowStarted
	go func() { defer wg.Done(); w.Check() }()
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if len(reported) == 0 || reported[len(reported)-1] != "dark" {
		t.Fatalf("reported %v; the last report must be the latest state (dark)", reported)
	}
}
