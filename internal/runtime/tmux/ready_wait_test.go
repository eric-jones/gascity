package tmux

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestAwaitRuntimeReady_PromptAppears verifies the happy path: once the ready
// prompt shows up in captured output, the wait returns nil.
func TestAwaitRuntimeReady_PromptAppears(t *testing.T) {
	calls := 0
	capture := func() ([]string, error) {
		calls++
		if calls >= 2 {
			return []string{"welcome", "❯ "}, nil
		}
		return []string{"starting..."}, nil
	}
	paneAlive := func() (bool, error) { return false, nil }

	if err := awaitRuntimeReady(context.Background(), 2*time.Second, "❯ ", capture, paneAlive); err != nil {
		t.Fatalf("awaitRuntimeReady = %v, want nil", err)
	}
}

// TestAwaitRuntimeReady_PaneDeathReturnsSentinelFast is the regression for
// tr-xg70y: when the agent process exits before readiness (e.g. a resume into a
// stale session key, kept visible by remain-on-exit), the wait must return
// errRuntimeDiedDuringReady promptly rather than polling the full timeout. The
// 64s "context deadline exceeded" resume hang was exactly this: a corpse pane
// never shows the ready prompt, so the wait burned the entire startup deadline.
func TestAwaitRuntimeReady_PaneDeathReturnsSentinelFast(t *testing.T) {
	capture := func() ([]string, error) { return []string{"No conversation found"}, nil }
	deadProbes := 0
	paneDead := func() (bool, error) {
		deadProbes++
		return deadProbes >= 2, nil
	}

	start := time.Now()
	err := awaitRuntimeReady(context.Background(), 30*time.Second, "❯ ", capture, paneDead)
	elapsed := time.Since(start)

	if !errors.Is(err, errRuntimeDiedDuringReady) {
		t.Fatalf("awaitRuntimeReady = %v, want errRuntimeDiedDuringReady", err)
	}
	if elapsed > 5*time.Second {
		t.Fatalf("awaitRuntimeReady took %v; expected fast death detection, not the full timeout", elapsed)
	}
}

// TestAwaitRuntimeReady_PaneDeadProbeErrorIsNonFatal verifies that a transient
// error from the pane-death probe (e.g. the pane is briefly unqueryable right
// after launch) does not abort the wait — a prompt that appears later still
// wins.
func TestAwaitRuntimeReady_PaneDeadProbeErrorIsNonFatal(t *testing.T) {
	calls := 0
	capture := func() ([]string, error) {
		calls++
		if calls >= 3 {
			return []string{"❯ "}, nil
		}
		return []string{"booting"}, nil
	}
	paneDead := func() (bool, error) { return false, errors.New("pane not queryable yet") }

	if err := awaitRuntimeReady(context.Background(), 2*time.Second, "❯ ", capture, paneDead); err != nil {
		t.Fatalf("awaitRuntimeReady = %v, want nil (probe error must be non-fatal)", err)
	}
}

// TestAwaitRuntimeReady_TimeoutWithoutPromptOrDeath verifies that a live pane
// that simply never shows the prompt still times out (best-effort), and that
// the timeout is distinct from the death sentinel.
func TestAwaitRuntimeReady_TimeoutWithoutPromptOrDeath(t *testing.T) {
	capture := func() ([]string, error) { return []string{"still working"}, nil }
	paneDead := func() (bool, error) { return false, nil }

	err := awaitRuntimeReady(context.Background(), 150*time.Millisecond, "❯ ", capture, paneDead)
	if err == nil {
		t.Fatal("awaitRuntimeReady = nil, want a timeout error")
	}
	if errors.Is(err, errRuntimeDiedDuringReady) {
		t.Fatalf("awaitRuntimeReady = %v, want a timeout error, not the death sentinel", err)
	}
}

// TestAwaitRuntimeReady_ContextCanceled verifies cancellation propagates.
func TestAwaitRuntimeReady_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	capture := func() ([]string, error) { return []string{"x"}, nil }
	paneDead := func() (bool, error) { return false, nil }

	if err := awaitRuntimeReady(ctx, 5*time.Second, "❯ ", capture, paneDead); !errors.Is(err, context.Canceled) {
		t.Fatalf("awaitRuntimeReady = %v, want context.Canceled", err)
	}
}
