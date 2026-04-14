package arbor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	arbor "github.com/ToySin/go-arbor"
)

func TestBlackboard_SetAndGet(t *testing.T) {
	bb := arbor.NewBlackboard()
	bb.Set("key", "value")

	v, ok := bb.Get("key")

	assert.True(t, ok)
	assert.Equal(t, "value", v)
}

func TestBlackboard_GetMissing(t *testing.T) {
	bb := arbor.NewBlackboard()

	v, ok := bb.Get("missing")

	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestBlackboard_Delete(t *testing.T) {
	bb := arbor.NewBlackboard()
	bb.Set("key", "value")
	bb.Delete("key")

	assert.False(t, bb.Has("key"))
}

func TestBlackboard_Has(t *testing.T) {
	bb := arbor.NewBlackboard()

	assert.False(t, bb.Has("key"))
	bb.Set("key", 42)
	assert.True(t, bb.Has("key"))
}

func TestBlackboard_Clear(t *testing.T) {
	bb := arbor.NewBlackboard()
	bb.Set("a", 1)
	bb.Set("b", 2)
	bb.Clear()

	assert.False(t, bb.Has("a"))
	assert.False(t, bb.Has("b"))
}

func TestGetTyped(t *testing.T) {
	bb := arbor.NewBlackboard()
	bb.Set("count", 42)
	bb.Set("name", "agent-1")

	count, ok := arbor.GetTyped[int](bb, "count")
	assert.True(t, ok)
	assert.Equal(t, 42, count)

	name, ok := arbor.GetTyped[string](bb, "name")
	assert.True(t, ok)
	assert.Equal(t, "agent-1", name)
}

func TestGetTyped_WrongType(t *testing.T) {
	bb := arbor.NewBlackboard()
	bb.Set("count", "not-an-int")

	v, ok := arbor.GetTyped[int](bb, "count")
	assert.False(t, ok)
	assert.Equal(t, 0, v)
}

func TestGetTyped_Missing(t *testing.T) {
	bb := arbor.NewBlackboard()

	v, ok := arbor.GetTyped[string](bb, "missing")
	assert.False(t, ok)
	assert.Equal(t, "", v)
}

func TestBlackboard_FromContext(t *testing.T) {
	bb := arbor.NewBlackboard()
	bb.Set("key", "value")

	ctx := arbor.WithBlackboard(context.Background(), bb)
	retrieved := arbor.BlackboardFrom(ctx)

	assert.NotNil(t, retrieved)
	v, ok := retrieved.Get("key")
	assert.True(t, ok)
	assert.Equal(t, "value", v)
}

func TestBlackboard_FromContext_Nil(t *testing.T) {
	bb := arbor.BlackboardFrom(context.Background())

	assert.Nil(t, bb)
}

func TestBlackboard_TreeIntegration(t *testing.T) {
	tree := arbor.NewTree(
		arbor.NewSequence("pipeline",
			arbor.NewAction("produce", func(ctx context.Context) arbor.Status {
				bb := arbor.BlackboardFrom(ctx)
				bb.Set("target", "agent-7")
				return arbor.Success
			}),
			arbor.NewAction("consume", func(ctx context.Context) arbor.Status {
				bb := arbor.BlackboardFrom(ctx)
				target, ok := arbor.GetTyped[string](bb, "target")
				if !ok || target != "agent-7" {
					return arbor.Failure
				}
				return arbor.Success
			}),
		),
	)

	status := tree.Tick(context.Background())

	assert.Equal(t, arbor.Success, status)
}

func TestBlackboard_PersistsAcrossTicks(t *testing.T) {
	tickCount := 0
	tree := arbor.NewTree(
		arbor.NewSequence("multi-tick",
			arbor.NewAction("write-once", func(ctx context.Context) arbor.Status {
				bb := arbor.BlackboardFrom(ctx)
				tickCount++
				if tickCount == 1 {
					bb.Set("data", "persisted")
				}
				return arbor.Success
			}),
			arbor.NewAction("read-always", func(ctx context.Context) arbor.Status {
				bb := arbor.BlackboardFrom(ctx)
				v, ok := arbor.GetTyped[string](bb, "data")
				if !ok || v != "persisted" {
					return arbor.Failure
				}
				return arbor.Success
			}),
		),
	)

	assert.Equal(t, arbor.Success, tree.Tick(context.Background()), "tick 1: write and read")
	assert.Equal(t, arbor.Success, tree.Tick(context.Background()), "tick 2: data persists")
}
