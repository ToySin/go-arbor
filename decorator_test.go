package arbor_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	arbor "github.com/ToySin/go-arbor"
)

// --- Inverter ---

func TestInverter_FlipsSuccess(t *testing.T) {
	inv := arbor.NewInverter("inv",
		arbor.NewAction("ok", func(ctx context.Context) arbor.Status { return arbor.Success }),
	)

	assert.Equal(t, arbor.Failure, inv.Tick(context.Background()))
}

func TestInverter_FlipsFailure(t *testing.T) {
	inv := arbor.NewInverter("inv",
		arbor.NewAction("fail", func(ctx context.Context) arbor.Status { return arbor.Failure }),
	)

	assert.Equal(t, arbor.Success, inv.Tick(context.Background()))
}

func TestInverter_RunningPassesThrough(t *testing.T) {
	inv := arbor.NewInverter("inv",
		arbor.NewAction("busy", func(ctx context.Context) arbor.Status { return arbor.Running }),
	)

	assert.Equal(t, arbor.Running, inv.Tick(context.Background()))
}

// --- Repeater ---

func TestRepeater_SucceedsAfterN(t *testing.T) {
	count := 0
	rep := arbor.NewRepeater("rep", 3,
		arbor.NewAction("work", func(ctx context.Context) arbor.Status {
			count++
			return arbor.Success
		}),
	)

	tree := arbor.NewTree(rep)

	assert.Equal(t, arbor.Running, tree.Tick(context.Background()), "tick 1")
	assert.Equal(t, arbor.Running, tree.Tick(context.Background()), "tick 2")
	assert.Equal(t, arbor.Success, tree.Tick(context.Background()), "tick 3")
	assert.Equal(t, 3, count)
}

func TestRepeater_FailsImmediately(t *testing.T) {
	count := 0
	rep := arbor.NewRepeater("rep", 5,
		arbor.NewAction("flaky", func(ctx context.Context) arbor.Status {
			count++
			if count == 2 {
				return arbor.Failure
			}
			return arbor.Success
		}),
	)

	tree := arbor.NewTree(rep)

	assert.Equal(t, arbor.Running, tree.Tick(context.Background()), "tick 1")
	assert.Equal(t, arbor.Failure, tree.Tick(context.Background()), "tick 2")
}

func TestRepeater_WaitsForRunning(t *testing.T) {
	tickCount := 0
	rep := arbor.NewRepeater("rep", 2,
		arbor.NewAction("slow", func(ctx context.Context) arbor.Status {
			tickCount++
			if tickCount%2 == 1 {
				return arbor.Running
			}
			return arbor.Success
		}),
	)

	tree := arbor.NewTree(rep)

	assert.Equal(t, arbor.Running, tree.Tick(context.Background()), "tick 1: child Running")
	assert.Equal(t, arbor.Running, tree.Tick(context.Background()), "tick 2: child Success, 1/2")
	assert.Equal(t, arbor.Running, tree.Tick(context.Background()), "tick 3: child Running")
	assert.Equal(t, arbor.Success, tree.Tick(context.Background()), "tick 4: child Success, 2/2")
}

// --- Retry ---

func TestRetry_SucceedsImmediately(t *testing.T) {
	r := arbor.NewRetry("retry", 3,
		arbor.NewAction("ok", func(ctx context.Context) arbor.Status { return arbor.Success }),
	)

	assert.Equal(t, arbor.Success, r.Tick(context.Background()))
}

func TestRetry_RetriesOnFailure(t *testing.T) {
	count := 0
	r := arbor.NewRetry("retry", 3,
		arbor.NewAction("flaky", func(ctx context.Context) arbor.Status {
			count++
			if count < 3 {
				return arbor.Failure
			}
			return arbor.Success
		}),
	)

	tree := arbor.NewTree(r)

	assert.Equal(t, arbor.Running, tree.Tick(context.Background()), "tick 1: fail, retry")
	assert.Equal(t, arbor.Running, tree.Tick(context.Background()), "tick 2: fail, retry")
	assert.Equal(t, arbor.Success, tree.Tick(context.Background()), "tick 3: success")
}

func TestRetry_ExhaustsRetries(t *testing.T) {
	r := arbor.NewRetry("retry", 2,
		arbor.NewAction("always-fail", func(ctx context.Context) arbor.Status { return arbor.Failure }),
	)

	tree := arbor.NewTree(r)

	assert.Equal(t, arbor.Running, tree.Tick(context.Background()), "tick 1")
	assert.Equal(t, arbor.Failure, tree.Tick(context.Background()), "tick 2: exhausted")
}

// --- Timeout ---

func TestTimeout_SucceedsWithinTime(t *testing.T) {
	to := arbor.NewTimeout("timeout", 1*time.Second,
		arbor.NewAction("fast", func(ctx context.Context) arbor.Status { return arbor.Success }),
	)

	assert.Equal(t, arbor.Success, to.Tick(context.Background()))
}

func TestTimeout_FailsAfterDuration(t *testing.T) {
	to := arbor.NewTimeout("timeout", 1*time.Millisecond,
		arbor.NewAction("slow", func(ctx context.Context) arbor.Status { return arbor.Running }),
	)

	tree := arbor.NewTree(to)

	assert.Equal(t, arbor.Running, tree.Tick(context.Background()), "tick 1: started")
	time.Sleep(2 * time.Millisecond)
	assert.Equal(t, arbor.Failure, tree.Tick(context.Background()), "tick 2: timed out")
}

// --- Visualization ---

func TestDecorator_Visualization(t *testing.T) {
	tree := arbor.NewTree(
		arbor.NewRetry("retry-connect", 3,
			arbor.NewInverter("not-blocked",
				arbor.NewCondition("is-blocked", func(ctx context.Context) bool {
					return true
				}),
			),
		),
	)

	tree.Tick(context.Background())
	output := arbor.SprintTree(tree)

	assert.Contains(t, output, "Retry: retry-connect")
	assert.Contains(t, output, "Inverter: not-blocked")
	assert.Contains(t, output, "Condition: is-blocked")
}
