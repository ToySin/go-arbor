package arbor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	arbor "github.com/ToySin/go-arbor"
)

// --- ReactiveSequence ---

func TestReactiveSequence_ReEvaluatesFromStart(t *testing.T) {
	condResult := true
	condTickCount := 0

	tree := arbor.NewTree(
		arbor.NewReactiveSequence("reactive",
			arbor.NewCondition("guard", func(ctx context.Context) bool {
				condTickCount++
				return condResult
			}),
			arbor.NewAction("work", func(ctx context.Context) arbor.Status {
				return arbor.Running
			}),
		),
	)

	// Tick 1: guard=Success, work=Running → Running
	assert.Equal(t, arbor.Running, tree.Tick(context.Background()))
	assert.Equal(t, 1, condTickCount)

	// Tick 2: guard re-evaluated (still true), work=Running
	assert.Equal(t, arbor.Running, tree.Tick(context.Background()))
	assert.Equal(t, 2, condTickCount, "guard should be re-evaluated every tick")
}

func TestReactiveSequence_HaltsOnConditionChange(t *testing.T) {
	condResult := true
	workHalted := false

	tree := arbor.NewTree(
		arbor.NewReactiveSequence("reactive",
			arbor.NewCondition("battery-ok", func(ctx context.Context) bool {
				return condResult
			}),
			arbor.NewAction("work",
				func(ctx context.Context) arbor.Status { return arbor.Running },
				arbor.WithHaltFunc(func() { workHalted = true }),
			),
		),
	)

	// Tick 1: condition true, work Running
	assert.Equal(t, arbor.Running, tree.Tick(context.Background()))
	assert.False(t, workHalted)

	// Condition changes
	condResult = false

	// Tick 2: condition fails → work should be halted → Failure
	assert.Equal(t, arbor.Failure, tree.Tick(context.Background()))
	assert.True(t, workHalted, "work should be halted when guard fails")
}

func TestReactiveSequence_AllSuccess(t *testing.T) {
	tree := arbor.NewTree(
		arbor.NewReactiveSequence("all-ok",
			arbor.NewCondition("check", func(ctx context.Context) bool { return true }),
			arbor.NewAction("work", func(ctx context.Context) arbor.Status { return arbor.Success }),
		),
	)

	assert.Equal(t, arbor.Success, tree.Tick(context.Background()))
}

func TestReactiveSequence_HaltsWhenDifferentChildRunning(t *testing.T) {
	tick := 0
	child1Halted := false

	tree := arbor.NewTree(
		arbor.NewReactiveSequence("reactive",
			arbor.NewAction("a1",
				func(ctx context.Context) arbor.Status {
					tick++
					if tick == 1 {
						return arbor.Running // Running on tick 1
					}
					return arbor.Success // Success on tick 2
				},
				arbor.WithHaltFunc(func() { child1Halted = true }),
			),
			arbor.NewAction("a2", func(ctx context.Context) arbor.Status {
				return arbor.Running
			}),
		),
	)

	// Tick 1: a1=Running → reactive Running, running=0
	assert.Equal(t, arbor.Running, tree.Tick(context.Background()))

	// Tick 2: a1=Success, a2=Running → running switches from 0 to 1 → a1 halted
	assert.Equal(t, arbor.Running, tree.Tick(context.Background()))
	assert.True(t, child1Halted, "a1 should be halted when running child changes")
}

// --- ReactiveFallback ---

func TestReactiveFallback_ReEvaluatesFromStart(t *testing.T) {
	condTickCount := 0

	tree := arbor.NewTree(
		arbor.NewReactiveFallback("reactive",
			arbor.NewCondition("check", func(ctx context.Context) bool {
				condTickCount++
				return false
			}),
			arbor.NewAction("fallback", func(ctx context.Context) arbor.Status {
				return arbor.Running
			}),
		),
	)

	tree.Tick(context.Background())
	tree.Tick(context.Background())

	assert.Equal(t, 2, condTickCount, "condition should be re-evaluated every tick")
}

func TestReactiveFallback_HigherPriorityPreempts(t *testing.T) {
	condResult := false
	fallbackHalted := false

	tree := arbor.NewTree(
		arbor.NewReactiveFallback("reactive",
			arbor.NewCondition("fast-path", func(ctx context.Context) bool {
				return condResult
			}),
			arbor.NewAction("slow-path",
				func(ctx context.Context) arbor.Status { return arbor.Running },
				arbor.WithHaltFunc(func() { fallbackHalted = true }),
			),
		),
	)

	// Tick 1: fast-path fails, slow-path Running
	assert.Equal(t, arbor.Running, tree.Tick(context.Background()))
	assert.False(t, fallbackHalted)

	// Fast-path now succeeds
	condResult = true

	// Tick 2: fast-path succeeds → slow-path halted → Success
	assert.Equal(t, arbor.Success, tree.Tick(context.Background()))
	assert.True(t, fallbackHalted, "slow-path should be halted when fast-path succeeds")
}

func TestReactiveFallback_AllFail(t *testing.T) {
	tree := arbor.NewTree(
		arbor.NewReactiveFallback("all-fail",
			arbor.NewCondition("c1", func(ctx context.Context) bool { return false }),
			arbor.NewCondition("c2", func(ctx context.Context) bool { return false }),
		),
	)

	assert.Equal(t, arbor.Failure, tree.Tick(context.Background()))
}

func TestReactiveFallback_FirstSucceeds(t *testing.T) {
	tree := arbor.NewTree(
		arbor.NewReactiveFallback("first-ok",
			arbor.NewCondition("check", func(ctx context.Context) bool { return true }),
			arbor.NewAction("never", func(ctx context.Context) arbor.Status {
				assert.Fail(t, "should not be reached")
				return arbor.Success
			}),
		),
	)

	assert.Equal(t, arbor.Success, tree.Tick(context.Background()))
}

// --- Visualization ---

func TestReactive_Visualization(t *testing.T) {
	tree := arbor.NewTree(
		arbor.NewReactiveSequence("guard",
			arbor.NewCondition("ok", func(ctx context.Context) bool { return true }),
			arbor.NewAction("work", func(ctx context.Context) arbor.Status { return arbor.Running }),
		),
	)

	tree.Tick(context.Background())
	output := arbor.SprintTree(tree)

	assert.Contains(t, output, "ReactiveSequence: guard (Running)")
	assert.Contains(t, output, "[✓] Condition: ok (Success)")
	assert.Contains(t, output, "[~] Action: work (Running)")
}
