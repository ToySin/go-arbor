# go-arbor

A generic Behavior Tree library for Go.

Tick-based execution following the standard BT formalism. Designed for job dispatchers, robot controllers, game AI, and any application that needs structured decision-making.

## Install

```bash
go get github.com/ToySin/go-arbor
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "os"

    arbor "github.com/ToySin/go-arbor"
)

func main() {
    tree := arbor.NewTree(
        arbor.NewFallback("root",
            arbor.NewSequence("try-primary",
                arbor.NewCondition("is-ready", func(ctx context.Context) bool {
                    return true
                }),
                arbor.NewAction("do-work", func(ctx context.Context) arbor.Status {
                    fmt.Println("working!")
                    return arbor.Success
                }),
            ),
            arbor.NewAction("fallback", func(ctx context.Context) arbor.Status {
                fmt.Println("fallback plan")
                return arbor.Success
            }),
        ),
    )

    tree.Tick(context.Background())
    arbor.PrintTree(os.Stdout, tree)
}
```

Output:

```
working!
[✓] Fallback: root (Success)
├── [✓] Sequence: try-primary (Success)
│   ├── [✓] Condition: is-ready (Success)
│   └── [✓] Action: do-work (Success)
└── [ ] Action: fallback (-)
```

## Node Types

### Composite — control flow with multiple children

| Node | Behavior |
|------|----------|
| `Sequence` | Ticks left-to-right. Fails on first failure. Succeeds when all succeed. |
| `Fallback` | Ticks left-to-right. Succeeds on first success. Fails when all fail. |
| `Parallel` | Ticks all children. Configurable success/failure threshold. |
| `ReactiveSequence` | Like Sequence, but re-evaluates from child 0 every tick. Halts previously Running children. |
| `ReactiveFallback` | Like Fallback, but re-evaluates from child 0 every tick. |

### Decorator — wraps a single child

| Node | Behavior |
|------|----------|
| `Inverter` | Flips Success ↔ Failure. Running passes through. |
| `Repeater` | Ticks child N times. Fails immediately on child failure. |
| `Retry` | Re-ticks child on failure, up to N attempts. |
| `Timeout` | Fails if child stays Running beyond the given duration. |

### Leaf — actual work

| Node | Behavior |
|------|----------|
| `Action` | Executes a function, returns Success / Failure / Running. |
| `Condition` | Evaluates a predicate, returns Success or Failure. Never Running. |

### Subtree — modular composition

Embed a tree as a node inside another tree, with isolated Blackboard and optional key mapping.

```go
innerTree := arbor.NewTree(arbor.NewAction("work", workFn))

mainTree := arbor.NewTree(
    arbor.NewSubtree("module", innerTree,
        arbor.WithInputMapping("parent_key", "subtree_key"),
        arbor.WithOutputMapping("subtree_result", "parent_result"),
    ),
)
```

## Blackboard

Shared key-value store for passing data between nodes. Automatically injected into context on each tick.

```go
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
            if !ok {
                return arbor.Failure
            }
            fmt.Println("assigned to", target)
            return arbor.Success
        }),
    ),
)
```

## Fluent Builder

Build trees with a chainable API:

```go
tree := arbor.NewBuilder().
    Sequence("dispatch").
        Condition("is-ready", readyFn).
        Retry("retry-assign", 3).
            Action("assign", assignFn).
        End().
        Action("notify", notifyFn).
    End().
    MustBuild()
```

## Tick Execution

```go
// Manual — caller controls when to tick
status := tree.Tick(ctx)

// Auto — tick loop at a fixed interval
tree.Run(ctx, 100*time.Millisecond,
    arbor.WithTickCallback(func(e arbor.TickEvent) bool {
        fmt.Printf("tick %d: %s\n", e.Tick, e.Status)
        return e.Status == arbor.Running // continue while Running
    }),
)
```

## Halt

Nodes can be interrupted while Running. Halt resets internal state and propagates down the tree.

```go
action := arbor.NewAction("work", workFn,
    arbor.WithHaltFunc(func() {
        fmt.Println("interrupted, cleaning up")
    }),
)
```

Reactive nodes (ReactiveSequence, ReactiveFallback) automatically halt previously Running children when re-evaluation changes the active branch.

## Visualization

```go
tree.Tick(ctx)
arbor.PrintTree(os.Stdout, tree)
// or
output := arbor.SprintTree(tree)
```

```
[~] Sequence: dispatch (Running)
├── [✓] Condition: agent-idle (Success)
├── [~] Retry: retry-assign (Running)
│   └── [✗] Action: assign-job (Failure)
└── [ ] Action: notify (-)
```

## Status

| Symbol | Status |
|--------|--------|
| `[✓]` | Success |
| `[✗]` | Failure |
| `[~]` | Running |
| `[ ]` | Not yet ticked |

## License

MIT
