package bt_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	bt "github.com/ToySin/go-bt"
)

func TestPrintTree_BeforeTick(t *testing.T) {
	tree := bt.NewTree(
		bt.NewSequence("root",
			bt.NewCondition("check", func(ctx context.Context) bool {
				return true
			}),
			bt.NewAction("work", func(ctx context.Context) bt.Status {
				return bt.Success
			}),
		),
	)

	output := bt.SprintTree(tree)

	assert.Contains(t, output, "Sequence: root (-)")
	assert.Contains(t, output, "Condition: check (-)")
	assert.Contains(t, output, "Action: work (-)")
	assert.Contains(t, output, "[ ]")
}

func TestPrintTree_AfterTick(t *testing.T) {
	tree := bt.NewTree(
		bt.NewSequence("root",
			bt.NewCondition("check", func(ctx context.Context) bool {
				return true
			}),
			bt.NewAction("work", func(ctx context.Context) bt.Status {
				return bt.Success
			}),
		),
	)

	tree.Tick(context.Background())
	output := bt.SprintTree(tree)

	assert.Contains(t, output, "[✓] Sequence: root (Success)")
	assert.Contains(t, output, "[✓] Condition: check (Success)")
	assert.Contains(t, output, "[✓] Action: work (Success)")
}

func TestPrintTree_Running(t *testing.T) {
	tree := bt.NewTree(
		bt.NewSequence("dispatch",
			bt.NewCondition("agent-idle", func(ctx context.Context) bool {
				return true
			}),
			bt.NewAction("assign-job", func(ctx context.Context) bt.Status {
				return bt.Running
			}),
			bt.NewAction("notify", func(ctx context.Context) bt.Status {
				return bt.Success
			}),
		),
	)

	tree.Tick(context.Background())
	output := bt.SprintTree(tree)

	assert.Contains(t, output, "[~] Sequence: dispatch (Running)")
	assert.Contains(t, output, "[✓] Condition: agent-idle (Success)")
	assert.Contains(t, output, "[~] Action: assign-job (Running)")
	assert.Contains(t, output, "[ ] Action: notify (-)")
}

func TestPrintTree_NestedFallback(t *testing.T) {
	tree := bt.NewTree(
		bt.NewFallback("root",
			bt.NewSequence("branch-1",
				bt.NewCondition("guard", func(ctx context.Context) bool {
					return false
				}),
				bt.NewAction("unreachable", func(ctx context.Context) bt.Status {
					return bt.Success
				}),
			),
			bt.NewAction("fallback-action", func(ctx context.Context) bt.Status {
				return bt.Success
			}),
		),
	)

	tree.Tick(context.Background())
	output := bt.SprintTree(tree)

	assert.Contains(t, output, "[✓] Fallback: root (Success)")
	assert.Contains(t, output, "[✗] Sequence: branch-1 (Failure)")
	assert.Contains(t, output, "[✗] Condition: guard (Failure)")
	assert.Contains(t, output, "[ ] Action: unreachable (-)")
	assert.Contains(t, output, "[✓] Action: fallback-action (Success)")
}
