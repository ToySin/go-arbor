package arbor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	arbor "github.com/ToySin/go-arbor"
)

func TestPrintTree_BeforeTick(t *testing.T) {
	tree := arbor.NewTree(
		arbor.NewSequence("root",
			arbor.NewCondition("check", func(ctx context.Context) bool {
				return true
			}),
			arbor.NewAction("work", func(ctx context.Context) arbor.Status {
				return arbor.Success
			}),
		),
	)

	output := arbor.SprintTree(tree)

	assert.Contains(t, output, "Sequence: root (-)")
	assert.Contains(t, output, "Condition: check (-)")
	assert.Contains(t, output, "Action: work (-)")
	assert.Contains(t, output, "[ ]")
}

func TestPrintTree_AfterTick(t *testing.T) {
	tree := arbor.NewTree(
		arbor.NewSequence("root",
			arbor.NewCondition("check", func(ctx context.Context) bool {
				return true
			}),
			arbor.NewAction("work", func(ctx context.Context) arbor.Status {
				return arbor.Success
			}),
		),
	)

	tree.Tick(context.Background())
	output := arbor.SprintTree(tree)

	assert.Contains(t, output, "[✓] Sequence: root (Success)")
	assert.Contains(t, output, "[✓] Condition: check (Success)")
	assert.Contains(t, output, "[✓] Action: work (Success)")
}

func TestPrintTree_Running(t *testing.T) {
	tree := arbor.NewTree(
		arbor.NewSequence("dispatch",
			arbor.NewCondition("agent-idle", func(ctx context.Context) bool {
				return true
			}),
			arbor.NewAction("assign-job", func(ctx context.Context) arbor.Status {
				return arbor.Running
			}),
			arbor.NewAction("notify", func(ctx context.Context) arbor.Status {
				return arbor.Success
			}),
		),
	)

	tree.Tick(context.Background())
	output := arbor.SprintTree(tree)

	assert.Contains(t, output, "[~] Sequence: dispatch (Running)")
	assert.Contains(t, output, "[✓] Condition: agent-idle (Success)")
	assert.Contains(t, output, "[~] Action: assign-job (Running)")
	assert.Contains(t, output, "[ ] Action: notify (-)")
}

func TestPrintTree_NestedFallback(t *testing.T) {
	tree := arbor.NewTree(
		arbor.NewFallback("root",
			arbor.NewSequence("branch-1",
				arbor.NewCondition("guard", func(ctx context.Context) bool {
					return false
				}),
				arbor.NewAction("unreachable", func(ctx context.Context) arbor.Status {
					return arbor.Success
				}),
			),
			arbor.NewAction("fallback-action", func(ctx context.Context) arbor.Status {
				return arbor.Success
			}),
		),
	)

	tree.Tick(context.Background())
	output := arbor.SprintTree(tree)

	assert.Contains(t, output, "[✓] Fallback: root (Success)")
	assert.Contains(t, output, "[✗] Sequence: branch-1 (Failure)")
	assert.Contains(t, output, "[✗] Condition: guard (Failure)")
	assert.Contains(t, output, "[ ] Action: unreachable (-)")
	assert.Contains(t, output, "[✓] Action: fallback-action (Success)")
}
