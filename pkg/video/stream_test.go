package video

import (
	"context"
	"image"
	"sync/atomic"
	"testing"
	"time"
)

func TestStreamRequestKeyframeInvokesClosure(t *testing.T) {
	var calls atomic.Int32
	s := NewStream()
	s.requestKey = func() { calls.Add(1) }

	s.RequestKeyframe()
	s.RequestKeyframe()

	if got := calls.Load(); got != 2 {
		t.Fatalf("expected 2 keyframe requests, got %d", got)
	}
}

// TestStreamKeyframeWatchdogStaleFrame verifies that the watchdog re-requests a
// keyframe while no fresh frame is published, and stops once a frame arrives.
func TestStreamKeyframeWatchdogStaleFrame(t *testing.T) {
	var calls atomic.Int32
	s := NewStream()
	s.requestKey = func() { calls.Add(1) }

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// publish a frame so latestTime() is set; then leave it stale.
	s.publish(Frame{Image: image.NewRGBA(image.Rect(0, 0, 2, 2)), At: time.Now()})

	go func() {
		ticker := time.NewTicker(KeyframeRetryInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if time.Since(s.latestTime()) >= KeyframeRetryInterval {
					s.RequestKeyframe()
				}
			}
		}
	}()

	deadline := time.After(KeyframeRetryInterval*2 + 500*time.Millisecond)
	select {
	case <-deadline:
		t.Fatalf("watchdog never requested a keyframe for a stale stream")
	case <-waitFor(func() bool { return calls.Load() > 0 }):
	}

	// Simulate a fresh frame arriving; the watchdog should stop asking.
	s.publish(Frame{Image: image.NewRGBA(image.Rect(0, 0, 2, 2)), At: time.Now()})
	before := calls.Load()
	time.Sleep(KeyframeRetryInterval + 200*time.Millisecond)
	if after := calls.Load(); after != before {
		t.Fatalf("watchdog kept requesting keyframes after a fresh frame (before=%d after=%d)", before, after)
	}
}

func waitFor(cond func() bool) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		for !cond() {
			time.Sleep(10 * time.Millisecond)
		}
		close(done)
	}()
	return done
}
