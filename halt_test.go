package arbor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	arbor "github.com/ToySin/go-arbor"
)

func TestSequence_Halt_ResetsCurrentChild(t *testing.T) {
	tickCount := 0
	seq := arbor.NewSequence("seq",
		arbor.NewAction("a1", func(ctx context.Context) arbor.Status {
			tickCount++
			return arbor.Success
		}),
		arbor.NewAction("a2", func(ctx context.Context) arbor.Status {
			return arbor.Running
		}),
	)

	tree := arbor.NewTree(seq)

	// Tick 1: a1 Success, a2 Running → Sequence Running at child 1
	assert.Equal(t, arbor.Running, tree.Tick(context.Background()))
	assert.Equal(t, 1, tickCount)

	// Halt → should reset to child 0
	seq.Halt()

	// Tick 2: should start from a1 again
	tree.Tick(context.Background())
	assert.Equal(t, 2, tickCount, "a1 should be ticked again after halt")
}

func TestFallback_Halt_ResetsCurrentChild(t *testing.T) {
	tickCount := 0
	fb := arbor.NewFallback("fb",
		arbor.NewAction("a1", func(ctx context.Context) arbor.Status {
			tickCount++
			return arbor.Running
		}),
		arbor.NewAction("a2", func(ctx context.Context) arbor.Status {
			return arbor.Success
		}),
	)

	tree := arbor.NewTree(fb)

	// Tick 1: a1 Running → Fallback Running at child 0
	assert.Equal(t, arbor.Running, tree.Tick(context.Background()))
	assert.Equal(t, 1, tickCount)

	// Halt → should reset
	fb.Halt()

	// Tick 2: should start from a1 again
	tree.Tick(context.Background())
	assert.Equal(t, 2, tickCount, "a1 should be ticked again after halt")
}

func TestHalt_PropagatesDownTree(t *testing.T) {
	innerHalted := false
	seq := arbor.NewSequence("outer",
		arbor.NewSequence("inner",
			arbor.NewAction("work", func(ctx context.Context) arbor.Status {
				return arbor.Running
			}, arbor.WithHaltFunc(func() {
				innerHalted = true
			})),
		),
	)

	tree := arbor.NewTree(seq)
	tree.Tick(context.Background())

	seq.Halt()

	assert.True(t, innerHalted, "halt should propagate to inner action")
}

func TestAction_HaltFunc_Called(t *testing.T) {
	halted := false
	action := arbor.NewAction("work",
		func(ctx context.Context) arbor.Status { return arbor.Running },
		arbor.WithHaltFunc(func() {
			halted = true
		}),
	)

	action.Tick(context.Background())
	action.Halt()

	assert.True(t, halted)
}

func TestAction_HaltFunc_NotSet(t *testing.T) {
	action := arbor.NewAction("work",
		func(ctx context.Context) arbor.Status { return arbor.Running },
	)

	action.Tick(context.Background())
	action.Halt() // should not panic

	assert.Nil(t, action.LastStatus())
}

func TestCondition_Halt_ResetsStatus(t *testing.T) {
	cond := arbor.NewCondition("check", func(ctx context.Context) bool {
		return true
	})

	cond.Tick(context.Background())
	assert.NotNil(t, cond.LastStatus())

	cond.Halt()
	assert.Nil(t, cond.LastStatus())
}

func TestParallel_Halt_HaltsAllRunningChildren(t *testing.T) {
	halt1 := false
	halt2 := false
	p := arbor.NewParallel("par", []arbor.Node{
		arbor.NewAction("a1",
			func(ctx context.Context) arbor.Status { return arbor.Running },
			arbor.WithHaltFunc(func() { halt1 = true }),
		),
		arbor.NewAction("a2",
			func(ctx context.Context) arbor.Status { return arbor.Running },
			arbor.WithHaltFunc(func() { halt2 = true }),
		),
	})

	tree := arbor.NewTree(p)
	tree.Tick(context.Background())

	p.Halt()

	assert.True(t, halt1, "a1 should be halted")
	assert.True(t, halt2, "a2 should be halted")
}

func TestDecorator_Halt_Propagates(t *testing.T) {
	halted := false

	inv := arbor.NewInverter("inv",
		arbor.NewAction("work",
			func(ctx context.Context) arbor.Status { return arbor.Running },
			arbor.WithHaltFunc(func() { halted = true }),
		),
	)
	inv.Tick(context.Background())
	inv.Halt()
	assert.True(t, halted, "inverter should propagate halt")
}

func TestRepeater_Halt_ResetsCounter(t *testing.T) {
	count := 0
	rep := arbor.NewRepeater("rep", 3,
		arbor.NewAction("work", func(ctx context.Context) arbor.Status {
			count++
			return arbor.Success
		}),
	)

	tree := arbor.NewTree(rep)
	tree.Tick(context.Background()) // count=1, internal=1/3 → Running
	tree.Tick(context.Background()) // count=2, internal=2/3 → Running

	rep.Halt() // reset internal counter to 0

	// After halt, needs 3 more successes (not 1)
	assert.Equal(t, arbor.Running, tree.Tick(context.Background()), "tick after halt: 1/3")
	assert.Equal(t, arbor.Running, tree.Tick(context.Background()), "tick after halt: 2/3")
	assert.Equal(t, arbor.Success, tree.Tick(context.Background()), "tick after halt: 3/3")
	assert.Equal(t, 5, count)
}

func TestRetry_Halt_ResetsAttempts(t *testing.T) {
	attempts := 0
	r := arbor.NewRetry("retry", 3,
		arbor.NewAction("flaky", func(ctx context.Context) arbor.Status {
			attempts++
			return arbor.Failure
		}),
	)

	tree := arbor.NewTree(r)
	tree.Tick(context.Background()) // attempt 1
	tree.Tick(context.Background()) // attempt 2

	r.Halt() // reset attempts

	// Should start fresh — 3 more failures needed to exhaust
	tree.Tick(context.Background()) // attempt 1 (reset)
	tree.Tick(context.Background()) // attempt 2
	status := tree.Tick(context.Background()) // attempt 3 → exhausted

	assert.Equal(t, arbor.Failure, status)
	assert.Equal(t, 5, attempts) // 2 before halt + 3 after
}
