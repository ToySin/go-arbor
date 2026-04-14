package arbor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	arbor "github.com/ToySin/go-arbor"
)

func TestParallel_DefaultPolicy_AllSucceed(t *testing.T) {
	p := arbor.NewParallel("all-succeed", []arbor.Node{
		arbor.NewAction("a1", func(ctx context.Context) arbor.Status { return arbor.Success }),
		arbor.NewAction("a2", func(ctx context.Context) arbor.Status { return arbor.Success }),
		arbor.NewAction("a3", func(ctx context.Context) arbor.Status { return arbor.Success }),
	})

	status := arbor.NewTree(p).Tick(context.Background())

	assert.Equal(t, arbor.Success, status)
}

func TestParallel_DefaultPolicy_OneFailure(t *testing.T) {
	p := arbor.NewParallel("fail-fast", []arbor.Node{
		arbor.NewAction("a1", func(ctx context.Context) arbor.Status { return arbor.Failure }),
		arbor.NewAction("a2", func(ctx context.Context) arbor.Status { return arbor.Success }),
		arbor.NewAction("a3", func(ctx context.Context) arbor.Status { return arbor.Success }),
	})

	status := arbor.NewTree(p).Tick(context.Background())

	assert.Equal(t, arbor.Failure, status)
}

func TestParallel_CustomThreshold(t *testing.T) {
	tickCount := 0
	p := arbor.NewParallel("gradual", []arbor.Node{
		arbor.NewAction("fast", func(ctx context.Context) arbor.Status {
			return arbor.Success
		}),
		arbor.NewAction("slow", func(ctx context.Context) arbor.Status {
			tickCount++
			if tickCount < 3 {
				return arbor.Running
			}
			return arbor.Success
		}),
	},
		arbor.WithSuccessThreshold(2),
		arbor.WithFailureThreshold(2),
	)

	tree := arbor.NewTree(p)

	assert.Equal(t, arbor.Running, tree.Tick(context.Background()), "tick 1")
	assert.Equal(t, arbor.Running, tree.Tick(context.Background()), "tick 2")
	assert.Equal(t, arbor.Success, tree.Tick(context.Background()), "tick 3")
	assert.Equal(t, 3, tickCount)
}

func TestParallel_SkipsCompletedChildren(t *testing.T) {
	a1TickCount := 0
	p := arbor.NewParallel("memory", []arbor.Node{
		arbor.NewAction("a1", func(ctx context.Context) arbor.Status {
			a1TickCount++
			return arbor.Success
		}),
		arbor.NewAction("a2", func(ctx context.Context) arbor.Status {
			return arbor.Running
		}),
	},
		arbor.WithSuccessThreshold(2),
		arbor.WithFailureThreshold(2),
	)

	tree := arbor.NewTree(p)
	tree.Tick(context.Background())
	tree.Tick(context.Background())

	assert.Equal(t, 1, a1TickCount, "a1 should only be ticked once")
}

func TestParallel_ResetsAfterCompletion(t *testing.T) {
	callCount := 0
	p := arbor.NewParallel("resettable", []arbor.Node{
		arbor.NewAction("a1", func(ctx context.Context) arbor.Status {
			callCount++
			return arbor.Success
		}),
		arbor.NewAction("a2", func(ctx context.Context) arbor.Status {
			return arbor.Success
		}),
	})

	tree := arbor.NewTree(p)
	tree.Tick(context.Background())
	tree.Tick(context.Background())

	assert.Equal(t, 2, callCount, "a1 should be ticked again after reset")
}

func TestParallel_SuccessOnOne(t *testing.T) {
	p := arbor.NewParallel("any-one", []arbor.Node{
		arbor.NewAction("a1", func(ctx context.Context) arbor.Status { return arbor.Running }),
		arbor.NewAction("a2", func(ctx context.Context) arbor.Status { return arbor.Success }),
		arbor.NewAction("a3", func(ctx context.Context) arbor.Status { return arbor.Running }),
	},
		arbor.WithSuccessThreshold(1),
		arbor.WithFailureThreshold(3),
	)

	status := arbor.NewTree(p).Tick(context.Background())

	assert.Equal(t, arbor.Success, status)
}

func TestParallel_Visualization(t *testing.T) {
	p := arbor.NewParallel("tasks", []arbor.Node{
		arbor.NewAction("fast", func(ctx context.Context) arbor.Status { return arbor.Success }),
		arbor.NewAction("slow", func(ctx context.Context) arbor.Status { return arbor.Running }),
	},
		arbor.WithSuccessThreshold(2),
		arbor.WithFailureThreshold(2),
	)

	tree := arbor.NewTree(p)
	tree.Tick(context.Background())
	output := arbor.SprintTree(tree)

	assert.Contains(t, output, "[~] Parallel: tasks (Running)")
	assert.Contains(t, output, "[✓] Action: fast (Success)")
	assert.Contains(t, output, "[~] Action: slow (Running)")
}
