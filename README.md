# go-arbor

A generic Behavior Tree library for Go.

Tick-based execution model following the standard BT formalism. Designed to be embedded in any Go application — from job dispatchers to robot controllers to game AI.

## Core Concepts

### Node Status

Every node returns one of three statuses on each tick:

- **Success** — node completed successfully
- **Failure** — node failed
- **Running** — node is still in progress, will resume on next tick

### Node Types

**Composite** — control flow with multiple children
- **Sequence** — ticks children left-to-right, fails on first failure, succeeds when all succeed
- **Fallback (Selector)** — ticks children left-to-right, succeeds on first success, fails when all fail
- **Parallel** — ticks all children concurrently, configurable success/failure policy

**Decorator** — wraps a single child, modifies behavior
- **Inverter** — flips Success ↔ Failure
- **Repeater** — re-ticks child N times or until failure
- **Retry** — re-ticks child on failure, up to N attempts
- **Timeout** — fails child if it stays Running beyond a duration

**Leaf** — actual work
- **Action** — executes logic, returns status
- **Condition** — evaluates a predicate, returns Success or Failure

### Blackboard

A shared key-value store accessible by all nodes in a tree. Used to pass data between nodes without coupling them directly.

### Tick Execution

```go
// Manual tick — caller controls when to tick
status := tree.Tick(ctx)

// Auto tick — library runs a tick loop at the given interval
tree.Run(ctx, 100*time.Millisecond)
```

## Phases

### Phase 1 — Core

- Node interface + Status (Success, Failure, Running)
- Composite nodes: Sequence, Fallback
- Leaf nodes: Action, Condition
- Tick-based execution with Running node re-entry

### Phase 2 — Advanced Nodes + Visualization

- Parallel composite node
- Decorator nodes: Inverter, Repeater, Retry, Timeout
- Blackboard (shared context between nodes)
- Tick runner (`tree.Run(ctx, interval)`)
- Tree structure visualization with live node status

### Phase 3 — Builder & Configuration

- Fluent tree builder API
- YAML/JSON tree definition and loading
- Tree structure validation at build time

### Phase 4 — Observability

- Per-tick logging and tracing
- Per-node execution statistics

## License

TBD
