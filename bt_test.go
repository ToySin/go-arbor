package bt_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	bt "github.com/ToySin/go-bt"
)

func TestStatus_String(t *testing.T) {
	assert.Equal(t, "Success", bt.Success.String())
	assert.Equal(t, "Failure", bt.Failure.String())
	assert.Equal(t, "Running", bt.Running.String())
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

	status := bt.NewTree(seq).Tick(context.Background())

	assert.Equal(t, bt.Success, status)
	assert.Equal(t, []string{"a1", "a2"}, called)
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

	assert.Equal(t, bt.Failure, status)
	assert.Equal(t, []string{"a1"}, called)
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

	assert.Equal(t, bt.Running, tree.Tick(context.Background()), "tick 1")
	assert.Equal(t, bt.Running, tree.Tick(context.Background()), "tick 2")
	assert.Equal(t, bt.Success, tree.Tick(context.Background()), "tick 3")
	assert.Equal(t, 3, tickCount)
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

	assert.Equal(t, bt.Success, status)
	assert.Equal(t, []string{"a1", "a2"}, called)
}

func TestFallback_AllFail(t *testing.T) {
	fb := bt.NewFallback("test-fb",
		bt.NewAction("a1", func(ctx context.Context) bt.Status { return bt.Failure }),
		bt.NewAction("a2", func(ctx context.Context) bt.Status { return bt.Failure }),
	)

	status := bt.NewTree(fb).Tick(context.Background())

	assert.Equal(t, bt.Failure, status)
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

	assert.Equal(t, bt.Running, tree.Tick(context.Background()), "tick 1")
	assert.Equal(t, bt.Success, tree.Tick(context.Background()), "tick 2")
}

func TestCondition(t *testing.T) {
	cond := bt.NewCondition("is-true", func(ctx context.Context) bool {
		return true
	})
	assert.Equal(t, bt.Success, cond.Tick(context.Background()))

	cond2 := bt.NewCondition("is-false", func(ctx context.Context) bool {
		return false
	})
	assert.Equal(t, bt.Failure, cond2.Tick(context.Background()))
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

	assert.Equal(t, bt.Success, status)
	assert.True(t, executed)
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

	assert.Equal(t, bt.Failure, status)
	assert.False(t, executed)
}

func TestNestedTree(t *testing.T) {
	// Fallback
	// ├── Sequence (fails because condition is false)
	// │   ├── Condition: false
	// │   └── Action: should not run
	// └── Action: fallback action (should run)
	executed := false
	tree := bt.NewTree(
		bt.NewFallback("root",
			bt.NewSequence("branch-1",
				bt.NewCondition("false-guard", func(ctx context.Context) bool {
					return false
				}),
				bt.NewAction("unreachable", func(ctx context.Context) bt.Status {
					assert.Fail(t, "should not be reached")
					return bt.Success
				}),
			),
			bt.NewAction("fallback-action", func(ctx context.Context) bt.Status {
				executed = true
				return bt.Success
			}),
		),
	)

	status := tree.Tick(context.Background())

	assert.Equal(t, bt.Success, status)
	assert.True(t, executed)
}
