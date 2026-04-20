package arbor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	arbor "github.com/ToySin/go-arbor"
)

func TestSubtree_DelegatesTickToInnerTree(t *testing.T) {
	executed := false
	inner := arbor.NewTree(
		arbor.NewAction("work", func(ctx context.Context) arbor.Status {
			executed = true
			return arbor.Success
		}),
	)

	tree := arbor.NewTree(
		arbor.NewSubtree("sub", inner),
	)

	status := tree.Tick(context.Background())

	assert.Equal(t, arbor.Success, status)
	assert.True(t, executed)
}

func TestSubtree_IsolatedBlackboard(t *testing.T) {
	inner := arbor.NewTree(
		arbor.NewAction("write", func(ctx context.Context) arbor.Status {
			bb := arbor.BlackboardFrom(ctx)
			bb.Set("inner_key", "inner_value")
			return arbor.Success
		}),
	)

	tree := arbor.NewTree(
		arbor.NewSequence("seq",
			arbor.NewSubtree("sub", inner),
			arbor.NewAction("check", func(ctx context.Context) arbor.Status {
				bb := arbor.BlackboardFrom(ctx)
				// inner_key should NOT be visible in parent BB
				if bb.Has("inner_key") {
					return arbor.Failure
				}
				return arbor.Success
			}),
		),
	)

	status := tree.Tick(context.Background())

	assert.Equal(t, arbor.Success, status)
}

func TestSubtree_InputMapping(t *testing.T) {
	inner := arbor.NewTree(
		arbor.NewAction("use-target", func(ctx context.Context) arbor.Status {
			bb := arbor.BlackboardFrom(ctx)
			target, ok := arbor.GetTyped[string](bb, "target")
			if !ok || target != "agent-7" {
				return arbor.Failure
			}
			return arbor.Success
		}),
	)

	tree := arbor.NewTree(
		arbor.NewSequence("seq",
			arbor.NewAction("set-target", func(ctx context.Context) arbor.Status {
				bb := arbor.BlackboardFrom(ctx)
				bb.Set("parent_target", "agent-7")
				return arbor.Success
			}),
			arbor.NewSubtree("sub", inner,
				arbor.WithInputMapping("parent_target", "target"),
			),
		),
	)

	status := tree.Tick(context.Background())

	assert.Equal(t, arbor.Success, status)
}

func TestSubtree_OutputMapping(t *testing.T) {
	inner := arbor.NewTree(
		arbor.NewAction("produce", func(ctx context.Context) arbor.Status {
			bb := arbor.BlackboardFrom(ctx)
			bb.Set("result", "done")
			return arbor.Success
		}),
	)

	tree := arbor.NewTree(
		arbor.NewSequence("seq",
			arbor.NewSubtree("sub", inner,
				arbor.WithOutputMapping("result", "parent_result"),
			),
			arbor.NewAction("check", func(ctx context.Context) arbor.Status {
				bb := arbor.BlackboardFrom(ctx)
				v, ok := arbor.GetTyped[string](bb, "parent_result")
				if !ok || v != "done" {
					return arbor.Failure
				}
				return arbor.Success
			}),
		),
	)

	status := tree.Tick(context.Background())

	assert.Equal(t, arbor.Success, status)
}

func TestSubtree_Running(t *testing.T) {
	tickCount := 0
	inner := arbor.NewTree(
		arbor.NewAction("slow", func(ctx context.Context) arbor.Status {
			tickCount++
			if tickCount < 3 {
				return arbor.Running
			}
			return arbor.Success
		}),
	)

	tree := arbor.NewTree(
		arbor.NewSubtree("sub", inner),
	)

	assert.Equal(t, arbor.Running, tree.Tick(context.Background()))
	assert.Equal(t, arbor.Running, tree.Tick(context.Background()))
	assert.Equal(t, arbor.Success, tree.Tick(context.Background()))
}

func TestSubtree_Halt(t *testing.T) {
	halted := false
	inner := arbor.NewTree(
		arbor.NewAction("work",
			func(ctx context.Context) arbor.Status { return arbor.Running },
			arbor.WithHaltFunc(func() { halted = true }),
		),
	)

	sub := arbor.NewSubtree("sub", inner)
	tree := arbor.NewTree(sub)
	tree.Tick(context.Background())

	sub.Halt()

	assert.True(t, halted)
	assert.Nil(t, sub.LastStatus())
}

func TestSubtree_Reuse(t *testing.T) {
	count := 0
	inner := arbor.NewTree(
		arbor.NewAction("work", func(ctx context.Context) arbor.Status {
			count++
			return arbor.Success
		}),
	)

	tree := arbor.NewTree(
		arbor.NewSequence("seq",
			arbor.NewSubtree("first", inner),
			arbor.NewSubtree("second", inner),
		),
	)

	tree.Tick(context.Background())

	assert.Equal(t, 2, count, "inner tree should be ticked twice (reused)")
}

func TestSubtree_Visualization(t *testing.T) {
	inner := arbor.NewTree(
		arbor.NewSequence("inner-seq",
			arbor.NewAction("work", func(ctx context.Context) arbor.Status {
				return arbor.Success
			}),
		),
	)

	tree := arbor.NewTree(
		arbor.NewSubtree("my-subtree", inner),
	)

	tree.Tick(context.Background())
	output := arbor.SprintTree(tree)

	assert.Contains(t, output, "Subtree: my-subtree (Success)")
	assert.Contains(t, output, "Sequence: inner-seq (Success)")
	assert.Contains(t, output, "Action: work (Success)")
}
