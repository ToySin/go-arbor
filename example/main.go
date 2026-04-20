package main

import (
	"context"
	"fmt"
	"os"
	"time"

	arbor "github.com/ToySin/go-arbor"
)

// Simulated agent state
var (
	agentIdle      = true
	batteryLevel   = 80
	assignAttempts = 0
)

func buildTree() *arbor.Tree {
	return arbor.NewTree(
		arbor.NewFallback("dispatch",
			// Primary path: find agent and assign job
			arbor.NewSequence("try-assign",
				// Check agent availability
				arbor.NewCondition("agent-idle", func(ctx context.Context) bool {
					return agentIdle
				}),
				// Check battery with inverter (fail if battery LOW)
				arbor.NewInverter("not-low-battery",
					arbor.NewCondition("battery-low", func(ctx context.Context) bool {
						return batteryLevel < 20
					}),
				),
				// Find best agent and store in blackboard
				arbor.NewAction("find-agent", func(ctx context.Context) arbor.Status {
					bb := arbor.BlackboardFrom(ctx)
					bb.Set("target_agent", "agent-7")
					bb.Set("job_id", "job-42")
					fmt.Println("  >> Found best agent: agent-7")
					return arbor.Success
				}),
				// Assign job with retry (simulate flaky network)
				arbor.NewRetry("retry-assign", 3,
					arbor.NewAction("assign-job", func(ctx context.Context) arbor.Status {
						bb := arbor.BlackboardFrom(ctx)
						agent, _ := arbor.GetTyped[string](bb, "target_agent")
						jobID, _ := arbor.GetTyped[string](bb, "job_id")
						assignAttempts++
						if assignAttempts < 2 {
							fmt.Printf("  >> Assign %s to %s failed (attempt %d)\n", jobID, agent, assignAttempts)
							return arbor.Failure
						}
						fmt.Printf("  >> Assigned %s to %s\n", jobID, agent)
						return arbor.Success
					}),
				),
				// Notify after assignment
				arbor.NewAction("notify", func(ctx context.Context) arbor.Status {
					bb := arbor.BlackboardFrom(ctx)
					agent, _ := arbor.GetTyped[string](bb, "target_agent")
					fmt.Printf("  >> Notified %s\n", agent)
					return arbor.Success
				}),
			),
			// Fallback: queue the job for later
			arbor.NewAction("queue-job", func(ctx context.Context) arbor.Status {
				fmt.Println("  >> No suitable agent, job queued")
				return arbor.Success
			}),
		),
	)
}

func main() {
	fmt.Println("========================================")
	fmt.Println(" go-arbor example: Job Dispatcher")
	fmt.Println("========================================")

	// --- Scenario 1: Normal dispatch with retry ---
	fmt.Println("\n--- Scenario 1: Normal dispatch (agent idle, battery ok) ---")
	agentIdle = true
	batteryLevel = 80
	assignAttempts = 0
	tree := buildTree()

	fmt.Println("\n[Tick 1]")
	tree.Tick(context.Background())
	arbor.PrintTree(os.Stdout, tree)

	fmt.Println("\n[Tick 2]")
	tree.Tick(context.Background())
	arbor.PrintTree(os.Stdout, tree)

	// --- Scenario 2: Low battery ---
	fmt.Println("\n--- Scenario 2: Low battery ---")
	batteryLevel = 10
	assignAttempts = 0
	tree = buildTree()

	fmt.Println("\n[Tick 1]")
	tree.Tick(context.Background())
	arbor.PrintTree(os.Stdout, tree)

	// --- Scenario 3: Agent busy ---
	fmt.Println("\n--- Scenario 3: Agent not idle ---")
	batteryLevel = 80
	agentIdle = false
	assignAttempts = 0
	tree = buildTree()

	fmt.Println("\n[Tick 1]")
	tree.Tick(context.Background())
	arbor.PrintTree(os.Stdout, tree)

	// --- Scenario 4: Build tree with fluent builder ---
	fmt.Println("\n--- Scenario 4: Fluent Builder API ---")
	assignAttempts = 0
	agentIdle = true
	batteryLevel = 80

	builderTree := arbor.NewBuilder().
		Fallback("dispatch").
		Sequence("try-assign").
		Condition("agent-idle", func(ctx context.Context) bool {
					return agentIdle
				}).
		Inverter("not-low-battery").
		Condition("battery-low", func(ctx context.Context) bool {
							return batteryLevel < 20
						}).
		End().
		Action("find-agent", func(ctx context.Context) arbor.Status {
					bb := arbor.BlackboardFrom(ctx)
					bb.Set("target_agent", "agent-7")
					bb.Set("job_id", "job-42")
					fmt.Println("  >> Found best agent: agent-7")
					return arbor.Success
				}).
		Retry("retry-assign", 3).
		Action("assign-job", func(ctx context.Context) arbor.Status {
						bb := arbor.BlackboardFrom(ctx)
						agent, _ := arbor.GetTyped[string](bb, "target_agent")
						jobID, _ := arbor.GetTyped[string](bb, "job_id")
						assignAttempts++
						if assignAttempts < 2 {
							fmt.Printf("  >> Assign %s to %s failed (attempt %d)\n", jobID, agent, assignAttempts)
							return arbor.Failure
						}
						fmt.Printf("  >> Assigned %s to %s\n", jobID, agent)
						return arbor.Success
					}).
		End().
		Action("notify", func(ctx context.Context) arbor.Status {
					bb := arbor.BlackboardFrom(ctx)
					agent, _ := arbor.GetTyped[string](bb, "target_agent")
					fmt.Printf("  >> Notified %s\n", agent)
					return arbor.Success
				}).
		End().
		Action("queue-job", func(ctx context.Context) arbor.Status {
				fmt.Println("  >> No suitable agent, job queued")
				return arbor.Success
			}).
		End().
		MustBuild()

	fmt.Println("\n[Tick 1]")
	builderTree.Tick(context.Background())
	arbor.PrintTree(os.Stdout, builderTree)

	fmt.Println("\n[Tick 2]")
	builderTree.Tick(context.Background())
	arbor.PrintTree(os.Stdout, builderTree)

	// --- Scenario 5: Auto tick with Run ---
	fmt.Println("\n--- Scenario 5: Auto tick with Run (500ms interval) ---")
	batteryLevel = 80
	agentIdle = true
	assignAttempts = 0
	tree = buildTree()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tree.Run(ctx, 500*time.Millisecond,
		arbor.WithTickCallback(func(e arbor.TickEvent) bool {
			fmt.Printf("\n[Auto Tick %d] status=%s\n", e.Tick, e.Status)
			arbor.PrintTree(os.Stdout, tree)
			// Stop after tree completes (no more Running nodes).
			return e.Status == arbor.Running
		}),
	)
}
