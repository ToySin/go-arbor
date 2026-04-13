package main

import (
	"context"
	"fmt"
	"os"

	bt "github.com/ToySin/go-bt"
)

func main() {
	batteryLevel := 80

	tree := bt.NewTree(
		bt.NewFallback("dispatch",
			bt.NewSequence("try-assign",
				bt.NewCondition("agent-idle", func(ctx context.Context) bool {
					return true
				}),
				bt.NewCondition("battery-ok", func(ctx context.Context) bool {
					return batteryLevel > 20
				}),
				bt.NewAction("assign-job", func(ctx context.Context) bt.Status {
					return bt.Running
				}),
				bt.NewAction("notify-agent", func(ctx context.Context) bt.Status {
					return bt.Success
				}),
			),
			bt.NewAction("queue-job", func(ctx context.Context) bt.Status {
				return bt.Success
			}),
		),
	)

	fmt.Println("=== Before tick ===")
	bt.PrintTree(os.Stdout, tree)

	fmt.Println("\n=== After tick 1 ===")
	tree.Tick(context.Background())
	bt.PrintTree(os.Stdout, tree)

	fmt.Println("\n=== After tick 2 (job completed) ===")
	// Simulate job completion on next tick
	tree = bt.NewTree(
		bt.NewFallback("dispatch",
			bt.NewSequence("try-assign",
				bt.NewCondition("agent-idle", func(ctx context.Context) bool {
					return true
				}),
				bt.NewCondition("battery-ok", func(ctx context.Context) bool {
					return batteryLevel > 20
				}),
				bt.NewAction("assign-job", func(ctx context.Context) bt.Status {
					return bt.Success
				}),
				bt.NewAction("notify-agent", func(ctx context.Context) bt.Status {
					return bt.Success
				}),
			),
			bt.NewAction("queue-job", func(ctx context.Context) bt.Status {
				return bt.Success
			}),
		),
	)
	tree.Tick(context.Background())
	bt.PrintTree(os.Stdout, tree)

	fmt.Println("\n=== Low battery scenario ===")
	batteryLevel = 10
	tree = bt.NewTree(
		bt.NewFallback("dispatch",
			bt.NewSequence("try-assign",
				bt.NewCondition("agent-idle", func(ctx context.Context) bool {
					return true
				}),
				bt.NewCondition("battery-ok", func(ctx context.Context) bool {
					return batteryLevel > 20
				}),
				bt.NewAction("assign-job", func(ctx context.Context) bt.Status {
					return bt.Success
				}),
				bt.NewAction("notify-agent", func(ctx context.Context) bt.Status {
					return bt.Success
				}),
			),
			bt.NewAction("queue-job", func(ctx context.Context) bt.Status {
				fmt.Println("  >> Job queued (no suitable agent)")
				return bt.Success
			}),
		),
	)
	tree.Tick(context.Background())
	bt.PrintTree(os.Stdout, tree)
}
