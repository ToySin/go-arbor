package arbor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	arbor "github.com/ToySin/go-arbor"
)

func TestStatus_String(t *testing.T) {
	assert.Equal(t, "Success", arbor.Success.String())
	assert.Equal(t, "Failure", arbor.Failure.String())
	assert.Equal(t, "Running", arbor.Running.String())
}

func TestSequence_AllSuccess(t *testing.T) {
	called := make([]string, 0)
	seq := arbor.NewSequence("test-seq",
		arbor.NewAction("a1", func(ctx context.Context) arbor.Status {
			called = append(called, "a1")
			return arbor.Success
		}),
		arbor.NewAction("a2", func(ctx context.Context) arbor.Status {
			called = append(called, "a2")
			return arbor.Success
		}),
	)

	status := arbor.NewTree(seq).Tick(context.Background())

	assert.Equal(t, arbor.Success, status)
	assert.Equal(t, []string{"a1", "a2"}, called)
}

func TestSequence_FailsOnFirstFailure(t *testing.T) {
	called := make([]string, 0)
	seq := arbor.NewSequence("test-seq",
		arbor.NewAction("a1", func(ctx context.Context) arbor.Status {
			called = append(called, "a1")
			return arbor.Failure
		}),
		arbor.NewAction("a2", func(ctx context.Context) arbor.Status {
			called = append(called, "a2")
			return arbor.Success
		}),
	)

	status := arbor.NewTree(seq).Tick(context.Background())

	assert.Equal(t, arbor.Failure, status)
	assert.Equal(t, []string{"a1"}, called)
}

func TestSequence_Running_ResumesFromRunningChild(t *testing.T) {
	tickCount := 0
	seq := arbor.NewSequence("test-seq",
		arbor.NewAction("a1", func(ctx context.Context) arbor.Status {
			return arbor.Success
		}),
		arbor.NewAction("a2", func(ctx context.Context) arbor.Status {
			tickCount++
			if tickCount < 3 {
				return arbor.Running
			}
			return arbor.Success
		}),
	)

	tree := arbor.NewTree(seq)

	assert.Equal(t, arbor.Running, tree.Tick(context.Background()), "tick 1")
	assert.Equal(t, arbor.Running, tree.Tick(context.Background()), "tick 2")
	assert.Equal(t, arbor.Success, tree.Tick(context.Background()), "tick 3")
	assert.Equal(t, 3, tickCount)
}

func TestFallback_SucceedsOnFirstSuccess(t *testing.T) {
	called := make([]string, 0)
	fb := arbor.NewFallback("test-fb",
		arbor.NewAction("a1", func(ctx context.Context) arbor.Status {
			called = append(called, "a1")
			return arbor.Failure
		}),
		arbor.NewAction("a2", func(ctx context.Context) arbor.Status {
			called = append(called, "a2")
			return arbor.Success
		}),
		arbor.NewAction("a3", func(ctx context.Context) arbor.Status {
			called = append(called, "a3")
			return arbor.Success
		}),
	)

	status := arbor.NewTree(fb).Tick(context.Background())

	assert.Equal(t, arbor.Success, status)
	assert.Equal(t, []string{"a1", "a2"}, called)
}

func TestFallback_AllFail(t *testing.T) {
	fb := arbor.NewFallback("test-fb",
		arbor.NewAction("a1", func(ctx context.Context) arbor.Status { return arbor.Failure }),
		arbor.NewAction("a2", func(ctx context.Context) arbor.Status { return arbor.Failure }),
	)

	status := arbor.NewTree(fb).Tick(context.Background())

	assert.Equal(t, arbor.Failure, status)
}

func TestFallback_Running_ResumesFromRunningChild(t *testing.T) {
	tickCount := 0
	fb := arbor.NewFallback("test-fb",
		arbor.NewAction("a1", func(ctx context.Context) arbor.Status {
			tickCount++
			if tickCount < 2 {
				return arbor.Running
			}
			return arbor.Success
		}),
		arbor.NewAction("a2", func(ctx context.Context) arbor.Status {
			return arbor.Success
		}),
	)

	tree := arbor.NewTree(fb)

	assert.Equal(t, arbor.Running, tree.Tick(context.Background()), "tick 1")
	assert.Equal(t, arbor.Success, tree.Tick(context.Background()), "tick 2")
}

func TestCondition(t *testing.T) {
	cond := arbor.NewCondition("is-true", func(ctx context.Context) bool {
		return true
	})
	assert.Equal(t, arbor.Success, cond.Tick(context.Background()))

	cond2 := arbor.NewCondition("is-false", func(ctx context.Context) bool {
		return false
	})
	assert.Equal(t, arbor.Failure, cond2.Tick(context.Background()))
}

func TestComposite_ConditionAndAction(t *testing.T) {
	executed := false
	tree := arbor.NewTree(
		arbor.NewSequence("guard-and-act",
			arbor.NewCondition("check", func(ctx context.Context) bool {
				return true
			}),
			arbor.NewAction("do-work", func(ctx context.Context) arbor.Status {
				executed = true
				return arbor.Success
			}),
		),
	)

	status := tree.Tick(context.Background())

	assert.Equal(t, arbor.Success, status)
	assert.True(t, executed)
}

func TestComposite_ConditionGuardBlocks(t *testing.T) {
	executed := false
	tree := arbor.NewTree(
		arbor.NewSequence("guard-and-act",
			arbor.NewCondition("check", func(ctx context.Context) bool {
				return false
			}),
			arbor.NewAction("do-work", func(ctx context.Context) arbor.Status {
				executed = true
				return arbor.Success
			}),
		),
	)

	status := tree.Tick(context.Background())

	assert.Equal(t, arbor.Failure, status)
	assert.False(t, executed)
}

func TestNestedTree(t *testing.T) {
	// Fallback
	// ├── Sequence (fails because condition is false)
	// │   ├── Condition: false
	// │   └── Action: should not run
	// └── Action: fallback action (should run)
	executed := false
	tree := arbor.NewTree(
		arbor.NewFallback("root",
			arbor.NewSequence("branch-1",
				arbor.NewCondition("false-guard", func(ctx context.Context) bool {
					return false
				}),
				arbor.NewAction("unreachable", func(ctx context.Context) arbor.Status {
					assert.Fail(t, "should not be reached")
					return arbor.Success
				}),
			),
			arbor.NewAction("fallback-action", func(ctx context.Context) arbor.Status {
				executed = true
				return arbor.Success
			}),
		),
	)

	status := tree.Tick(context.Background())

	assert.Equal(t, arbor.Success, status)
	assert.True(t, executed)
}
