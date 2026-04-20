package arbor_test

import (
	"context"
	"testing"
	"time"

	arbor "github.com/ToySin/go-arbor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_IssueExample(t *testing.T) {
	isIdle := func(ctx context.Context) bool { return true }
	assignJob := func(ctx context.Context) arbor.Status { return arbor.Success }
	notify := func(ctx context.Context) arbor.Status { return arbor.Success }

	tree, err := arbor.NewBuilder().
		Sequence("dispatch").
		Condition("agent-idle", isIdle).
		Action("assign-job", assignJob).
		Action("notify", notify).
		End().
		Build()

	require.NoError(t, err)
	assert.Equal(t, arbor.Success, tree.Tick(context.Background()))
}

func TestBuilder_NestedComposites(t *testing.T) {
	tree, err := arbor.NewBuilder().
		Fallback("root").
		Sequence("try-first").
		Action("fail", func(ctx context.Context) arbor.Status {
				return arbor.Failure
			}).
		End().
		Action("fallback-action", func(ctx context.Context) arbor.Status {
				return arbor.Success
			}).
		End().
		Build()

	require.NoError(t, err)
	assert.Equal(t, arbor.Success, tree.Tick(context.Background()))
}

func TestBuilder_DecoratorWrapsLeaf(t *testing.T) {
	tree, err := arbor.NewBuilder().
		Inverter("flip").
		Action("fail", func(ctx context.Context) arbor.Status {
				return arbor.Failure
			}).
		End().
		Build()

	require.NoError(t, err)
	assert.Equal(t, arbor.Success, tree.Tick(context.Background()))
}

func TestBuilder_DecoratorWrapsComposite(t *testing.T) {
	tree, err := arbor.NewBuilder().
		Inverter("flip").
		Sequence("seq").
		Action("fail", func(ctx context.Context) arbor.Status {
					return arbor.Failure
				}).
		End().
		End().
		Build()

	require.NoError(t, err)
	assert.Equal(t, arbor.Success, tree.Tick(context.Background()))
}

func TestBuilder_Retry(t *testing.T) {
	attempts := 0
	tree, err := arbor.NewBuilder().
		Retry("retry", 3).
		Action("flaky", func(ctx context.Context) arbor.Status {
				attempts++
				if attempts < 3 {
					return arbor.Failure
				}
				return arbor.Success
			}).
		End().
		Build()

	require.NoError(t, err)
	// Retry ticks the child once per tree tick, retrying on failure.
	for tree.Tick(context.Background()) == arbor.Running {
	}
	assert.Equal(t, 3, attempts)
}

func TestBuilder_Repeater(t *testing.T) {
	count := 0
	tree, err := arbor.NewBuilder().
		Repeater("repeat", 3).
		Action("inc", func(ctx context.Context) arbor.Status {
				count++
				return arbor.Success
			}).
		End().
		Build()

	require.NoError(t, err)
	for tree.Tick(context.Background()) == arbor.Running {
	}
	assert.Equal(t, 3, count)
}

func TestBuilder_Timeout(t *testing.T) {
	tree, err := arbor.NewBuilder().
		Timeout("quick", 10*time.Millisecond).
		Action("slow", func(ctx context.Context) arbor.Status {
				return arbor.Running
			}).
		End().
		Build()

	require.NoError(t, err)
	// First tick starts the timer.
	assert.Equal(t, arbor.Running, tree.Tick(context.Background()))
	// After timeout expires, should fail.
	time.Sleep(15 * time.Millisecond)
	assert.Equal(t, arbor.Failure, tree.Tick(context.Background()))
}

func TestBuilder_Parallel(t *testing.T) {
	tree, err := arbor.NewBuilder().
		Parallel("all", arbor.WithSuccessThreshold(2)).
		Action("a1", func(ctx context.Context) arbor.Status { return arbor.Success }).
		Action("a2", func(ctx context.Context) arbor.Status { return arbor.Success }).
		Action("a3", func(ctx context.Context) arbor.Status { return arbor.Running }).
		End().
		Build()

	require.NoError(t, err)
	assert.Equal(t, arbor.Success, tree.Tick(context.Background()))
}

func TestBuilder_Error_UnclosedScope(t *testing.T) {
	_, err := arbor.NewBuilder().
		Sequence("open").
		Action("a", func(ctx context.Context) arbor.Status { return arbor.Success }).
		Build()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unclosed scope")
}

func TestBuilder_Error_ExtraEnd(t *testing.T) {
	_, err := arbor.NewBuilder().
		Sequence("s").
		Action("a", func(ctx context.Context) arbor.Status { return arbor.Success }).
		End().
		End().
		Build()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "without matching open scope")
}

func TestBuilder_Error_EmptyComposite(t *testing.T) {
	_, err := arbor.NewBuilder().
		Sequence("empty").
		End().
		Build()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no children")
}

func TestBuilder_Error_DecoratorTooManyChildren(t *testing.T) {
	_, err := arbor.NewBuilder().
		Inverter("flip").
		Action("a1", func(ctx context.Context) arbor.Status { return arbor.Success }).
		Action("a2", func(ctx context.Context) arbor.Status { return arbor.Success }).
		End().
		Build()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exactly 1 child")
}

func TestBuilder_Error_DecoratorNoChild(t *testing.T) {
	_, err := arbor.NewBuilder().
		Inverter("flip").
		End().
		Build()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exactly 1 child")
}

func TestBuilder_Error_EmptyTree(t *testing.T) {
	_, err := arbor.NewBuilder().Build()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exactly 1 root node, got 0")
}

func TestBuilder_MustBuild_Panics(t *testing.T) {
	assert.Panics(t, func() {
		arbor.NewBuilder().MustBuild()
	})
}

func TestBuilder_MustBuild_Success(t *testing.T) {
	assert.NotPanics(t, func() {
		tree := arbor.NewBuilder().
			Action("ok", func(ctx context.Context) arbor.Status { return arbor.Success }).
			MustBuild()
		assert.Equal(t, arbor.Success, tree.Tick(context.Background()))
	})
}

func TestBuilder_EquivalentToManual(t *testing.T) {
	fn := func(ctx context.Context) arbor.Status { return arbor.Success }
	cond := func(ctx context.Context) bool { return true }

	// Build with builder.
	builderTree, err := arbor.NewBuilder().
		Fallback("root").
		Sequence("path-a").
		Condition("check", cond).
		Inverter("flip").
		Condition("neg-check", cond).
		End().
		Action("do", fn).
		End().
		Action("path-b", fn).
		End().
		Build()
	require.NoError(t, err)

	// Build manually.
	manualTree := arbor.NewTree(
		arbor.NewFallback("root",
			arbor.NewSequence("path-a",
				arbor.NewCondition("check", cond),
				arbor.NewInverter("flip",
					arbor.NewCondition("neg-check", cond),
				),
				arbor.NewAction("do", fn),
			),
			arbor.NewAction("path-b", fn),
		),
	)

	// Both should produce same tick result.
	ctx := context.Background()
	assert.Equal(t, manualTree.Tick(ctx), builderTree.Tick(ctx))
}
