package bt_test

import (
	"context"
	"testing"

	bt "github.com/ToySin/go-bt"
)

func TestStatus_String(t *testing.T) {
	tests := []struct {
		status bt.Status
		want   string
	}{
		{bt.Success, "Success"},
		{bt.Failure, "Failure"},
		{bt.Running, "Running"},
	}
	for _, tt := range tests {
		if got := tt.status.String(); got != tt.want {
			t.Errorf("Status.String() = %q, want %q", got, tt.want)
		}
	}
}

func TestSequence_AllSuccess(t *testing.T) {
	called := make([]string, 0)
	seq := bt.NewSequence("test-seq",
		bt.NewAction("a1", func(ctx context.Context) bt.Status {
			called = append(called, "a1")
			return bt.Success
		}),
		bt.NewAction("a2", func(ctx context.Context) bt.Status {
			called = append(called, "a2")
			return bt.Success
		}),
	)

	tree := bt.NewTree(seq)
	status := tree.Tick(context.Background())

	if status != bt.Success {
		t.Errorf("got %v, want Success", status)
	}
	if len(called) != 2 || called[0] != "a1" || called[1] != "a2" {
		t.Errorf("expected [a1 a2], got %v", called)
	}
}

func TestSequence_FailsOnFirstFailure(t *testing.T) {
	called := make([]string, 0)
	seq := bt.NewSequence("test-seq",
		bt.NewAction("a1", func(ctx context.Context) bt.Status {
			called = append(called, "a1")
			return bt.Failure
		}),
		bt.NewAction("a2", func(ctx context.Context) bt.Status {
			called = append(called, "a2")
			return bt.Success
		}),
	)

	status := bt.NewTree(seq).Tick(context.Background())

	if status != bt.Failure {
		t.Errorf("got %v, want Failure", status)
	}
	if len(called) != 1 {
		t.Errorf("a2 should not have been called, got %v", called)
	}
}

func TestSequence_Running_ResumesFromRunningChild(t *testing.T) {
	tickCount := 0
	seq := bt.NewSequence("test-seq",
		bt.NewAction("a1", func(ctx context.Context) bt.Status {
			return bt.Success
		}),
		bt.NewAction("a2", func(ctx context.Context) bt.Status {
			tickCount++
			if tickCount < 3 {
				return bt.Running
			}
			return bt.Success
		}),
	)

	tree := bt.NewTree(seq)

	// First tick: a1 succeeds, a2 returns Running
	if s := tree.Tick(context.Background()); s != bt.Running {
		t.Errorf("tick 1: got %v, want Running", s)
	}
	// Second tick: resumes from a2, still Running
	if s := tree.Tick(context.Background()); s != bt.Running {
		t.Errorf("tick 2: got %v, want Running", s)
	}
	// Third tick: a2 succeeds
	if s := tree.Tick(context.Background()); s != bt.Success {
		t.Errorf("tick 3: got %v, want Success", s)
	}
	if tickCount != 3 {
		t.Errorf("a2 should have been ticked 3 times, got %d", tickCount)
	}
}

func TestFallback_SucceedsOnFirstSuccess(t *testing.T) {
	called := make([]string, 0)
	fb := bt.NewFallback("test-fb",
		bt.NewAction("a1", func(ctx context.Context) bt.Status {
			called = append(called, "a1")
			return bt.Failure
		}),
		bt.NewAction("a2", func(ctx context.Context) bt.Status {
			called = append(called, "a2")
			return bt.Success
		}),
		bt.NewAction("a3", func(ctx context.Context) bt.Status {
			called = append(called, "a3")
			return bt.Success
		}),
	)

	status := bt.NewTree(fb).Tick(context.Background())

	if status != bt.Success {
		t.Errorf("got %v, want Success", status)
	}
	if len(called) != 2 || called[1] != "a2" {
		t.Errorf("expected [a1 a2], got %v", called)
	}
}

func TestFallback_AllFail(t *testing.T) {
	fb := bt.NewFallback("test-fb",
		bt.NewAction("a1", func(ctx context.Context) bt.Status { return bt.Failure }),
		bt.NewAction("a2", func(ctx context.Context) bt.Status { return bt.Failure }),
	)

	status := bt.NewTree(fb).Tick(context.Background())

	if status != bt.Failure {
		t.Errorf("got %v, want Failure", status)
	}
}

func TestFallback_Running_ResumesFromRunningChild(t *testing.T) {
	tickCount := 0
	fb := bt.NewFallback("test-fb",
		bt.NewAction("a1", func(ctx context.Context) bt.Status {
			tickCount++
			if tickCount < 2 {
				return bt.Running
			}
			return bt.Success
		}),
		bt.NewAction("a2", func(ctx context.Context) bt.Status {
			return bt.Success
		}),
	)

	tree := bt.NewTree(fb)

	if s := tree.Tick(context.Background()); s != bt.Running {
		t.Errorf("tick 1: got %v, want Running", s)
	}
	if s := tree.Tick(context.Background()); s != bt.Success {
		t.Errorf("tick 2: got %v, want Success", s)
	}
}

func TestCondition(t *testing.T) {
	cond := bt.NewCondition("is-true", func(ctx context.Context) bool {
		return true
	})
	if s := cond.Tick(context.Background()); s != bt.Success {
		t.Errorf("got %v, want Success", s)
	}

	cond2 := bt.NewCondition("is-false", func(ctx context.Context) bool {
		return false
	})
	if s := cond2.Tick(context.Background()); s != bt.Failure {
		t.Errorf("got %v, want Failure", s)
	}
}

func TestComposite_ConditionAndAction(t *testing.T) {
	executed := false
	tree := bt.NewTree(
		bt.NewSequence("guard-and-act",
			bt.NewCondition("check", func(ctx context.Context) bool {
				return true
			}),
			bt.NewAction("do-work", func(ctx context.Context) bt.Status {
				executed = true
				return bt.Success
			}),
		),
	)

	status := tree.Tick(context.Background())

	if status != bt.Success {
		t.Errorf("got %v, want Success", status)
	}
	if !executed {
		t.Error("action should have been executed")
	}
}

func TestComposite_ConditionGuardBlocks(t *testing.T) {
	executed := false
	tree := bt.NewTree(
		bt.NewSequence("guard-and-act",
			bt.NewCondition("check", func(ctx context.Context) bool {
				return false
			}),
			bt.NewAction("do-work", func(ctx context.Context) bt.Status {
				executed = true
				return bt.Success
			}),
		),
	)

	status := tree.Tick(context.Background())

	if status != bt.Failure {
		t.Errorf("got %v, want Failure", status)
	}
	if executed {
		t.Error("action should not have been executed")
	}
}

func TestNestedTree(t *testing.T) {
	// Fallback
	// ├── Sequence (fails because condition is false)
	// │   ├── Condition: false
	// │   └── Action: should not run
	// └── Action: fallback action (should run)
	fallbackRan := false
	tree := bt.NewTree(
		bt.NewFallback("root",
			bt.NewSequence("branch-1",
				bt.NewCondition("false-guard", func(ctx context.Context) bool {
					return false
				}),
				bt.NewAction("unreachable", func(ctx context.Context) bt.Status {
					t.Error("should not be reached")
					return bt.Success
				}),
			),
			bt.NewAction("fallback-action", func(ctx context.Context) bt.Status {
				fallbackRan = true
				return bt.Success
			}),
		),
	)

	status := tree.Tick(context.Background())

	if status != bt.Success {
		t.Errorf("got %v, want Success", status)
	}
	if !fallbackRan {
		t.Error("fallback action should have run")
	}
}
