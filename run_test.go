package arbor_test

import (
	"context"
	"sync"
	"testing"
	"time"

	arbor "github.com/ToySin/go-arbor"
	"github.com/stretchr/testify/assert"
)

func TestRun_RespectsContextCancellation(t *testing.T) {
	var mu sync.Mutex
	tickCount := 0

	tree := arbor.NewTree(arbor.NewAction("inc", func(ctx context.Context) arbor.Status {
		mu.Lock()
		tickCount++
		mu.Unlock()
		return arbor.Success
	}))

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel after enough time for at least one tick.
	time.AfterFunc(50*time.Millisecond, cancel)

	err := tree.Run(ctx, 1*time.Millisecond)
	assert.ErrorIs(t, err, context.Canceled)

	mu.Lock()
	assert.GreaterOrEqual(t, tickCount, 1)
	mu.Unlock()
}

func TestRun_CallbackReceivesTickEvents(t *testing.T) {
	var events []arbor.TickEvent

	tree := arbor.NewTree(arbor.NewAction("ok", func(ctx context.Context) arbor.Status {
		return arbor.Success
	}))

	err := tree.Run(context.Background(), 1*time.Millisecond,
		arbor.WithTickCallback(func(e arbor.TickEvent) bool {
			events = append(events, e)
			return len(events) < 5
		}),
	)

	assert.NoError(t, err)
	assert.Len(t, events, 5)
	for i, e := range events {
		assert.Equal(t, i+1, e.Tick)
		assert.Equal(t, arbor.Success, e.Status)
	}
}

func TestRun_CallbackStopsLoop(t *testing.T) {
	tickCount := 0

	tree := arbor.NewTree(arbor.NewAction("ok", func(ctx context.Context) arbor.Status {
		tickCount++
		return arbor.Success
	}))

	err := tree.Run(context.Background(), 1*time.Millisecond,
		arbor.WithTickCallback(func(e arbor.TickEvent) bool {
			return e.Tick < 3
		}),
	)

	assert.NoError(t, err)
	assert.Equal(t, 3, tickCount)
}

func TestRun_NoCallback(t *testing.T) {
	tree := arbor.NewTree(arbor.NewAction("ok", func(ctx context.Context) arbor.Status {
		return arbor.Success
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := tree.Run(ctx, 1*time.Millisecond)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestRun_InvalidInterval(t *testing.T) {
	tree := arbor.NewTree(arbor.NewAction("ok", func(ctx context.Context) arbor.Status {
		return arbor.Success
	}))

	err := tree.Run(context.Background(), 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interval must be positive")

	err = tree.Run(context.Background(), -1*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interval must be positive")
}

func TestRun_RunningNodeResumesAcrossTicks(t *testing.T) {
	tickCount := 0
	seq := arbor.NewSequence("resume-seq",
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
	var events []arbor.TickEvent

	err := tree.Run(context.Background(), 1*time.Millisecond,
		arbor.WithTickCallback(func(e arbor.TickEvent) bool {
			events = append(events, e)
			// Stop after the sequence completes.
			return e.Status != arbor.Success
		}),
	)

	assert.NoError(t, err)
	assert.Len(t, events, 3)
	assert.Equal(t, arbor.Running, events[0].Status)
	assert.Equal(t, arbor.Running, events[1].Status)
	assert.Equal(t, arbor.Success, events[2].Status)
}
