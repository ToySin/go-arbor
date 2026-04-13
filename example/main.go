package main

import (
	"context"
	"fmt"
	"os"

	arbor "github.com/ToySin/go-arbor"
)

func main() {
	batteryLevel := 80

	tree := arbor.NewTree(
		arbor.NewFallback("dispatch",
			arbor.NewSequence("try-assign",
				arbor.NewCondition("agent-idle", func(ctx context.Context) bool {
					return true
				}),
				arbor.NewCondition("battery-ok", func(ctx context.Context) bool {
					return batteryLevel > 20
				}),
				arbor.NewAction("assign-job", func(ctx context.Context) arbor.Status {
					return arbor.Running
				}),
				arbor.NewAction("notify-agent", func(ctx context.Context) arbor.Status {
					return arbor.Success
				}),
			),
			arbor.NewAction("queue-job", func(ctx context.Context) arbor.Status {
				return arbor.Success
			}),
		),
	)

	fmt.Println("=== Before tick ===")
	arbor.PrintTree(os.Stdout, tree)

	fmt.Println("\n=== After tick 1 ===")
	tree.Tick(context.Background())
	arbor.PrintTree(os.Stdout, tree)

	fmt.Println("\n=== After tick 2 (job completed) ===")
	// Simulate job completion on next tick
	tree = arbor.NewTree(
		arbor.NewFallback("dispatch",
			arbor.NewSequence("try-assign",
				arbor.NewCondition("agent-idle", func(ctx context.Context) bool {
					return true
				}),
				arbor.NewCondition("battery-ok", func(ctx context.Context) bool {
					return batteryLevel > 20
				}),
				arbor.NewAction("assign-job", func(ctx context.Context) arbor.Status {
					return arbor.Success
				}),
				arbor.NewAction("notify-agent", func(ctx context.Context) arbor.Status {
					return arbor.Success
				}),
			),
			arbor.NewAction("queue-job", func(ctx context.Context) arbor.Status {
				return arbor.Success
			}),
		),
	)
	tree.Tick(context.Background())
	arbor.PrintTree(os.Stdout, tree)

	fmt.Println("\n=== Low battery scenario ===")
	batteryLevel = 10
	tree = arbor.NewTree(
		arbor.NewFallback("dispatch",
			arbor.NewSequence("try-assign",
				arbor.NewCondition("agent-idle", func(ctx context.Context) bool {
					return true
				}),
				arbor.NewCondition("battery-ok", func(ctx context.Context) bool {
					return batteryLevel > 20
				}),
				arbor.NewAction("assign-job", func(ctx context.Context) arbor.Status {
					return arbor.Success
				}),
				arbor.NewAction("notify-agent", func(ctx context.Context) arbor.Status {
					return arbor.Success
				}),
			),
			arbor.NewAction("queue-job", func(ctx context.Context) arbor.Status {
				fmt.Println("  >> Job queued (no suitable agent)")
				return arbor.Success
			}),
		),
	)
	tree.Tick(context.Background())
	arbor.PrintTree(os.Stdout, tree)
}
