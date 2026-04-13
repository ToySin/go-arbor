# Design Decisions & Discussion Log

Questions, concerns, and decisions made during the initial design and implementation of go-bt and Job-Dispatcher.

---

## 1. State Machine vs Behavior Tree

**Question:** Should the dispatcher engine use a state machine or behavior tree?

**Context:** State machine works well for Job lifecycle (PENDING → ASSIGNED → RUNNING → ...), but the dispatching engine's decision flow (filter → score → select → send → handle result) has deep conditional branching and fallback paths.

**Decision:** Both.
- Job lifecycle → State machine (clear, linear state transitions)
- Dispatcher engine → Behavior tree (flexible branching, fallback, easy to extend without modifying existing flow)

**Rationale:** State machine suffers from state explosion when branching gets deep. BT represents fallback/retry as tree structure naturally, and aligns well with the Strategy interface and YAML configuration goals.

---

## 2. Tick-based vs Event-driven Execution

**Question:** Go is typically event-driven. Is tick-based polling appropriate?

**Concern:** As a Go developer, periodic polling feels wasteful — a pattern usually avoided in server applications.

**Decision:** Tick-based, following standard BT formalism.

**Rationale:** Tick is not just polling — it's the re-evaluation mechanism for `Running` nodes. When a node returns Running, the next tick checks whether it's still running and whether the broader tree context has changed (e.g., higher-priority branch should preempt). Event-driven can miss cross-tree state changes. The library is generic, so following the established BT model is more important than optimizing for one use case. Users control tick interval or can tick manually on events.

---

## 3. ctx Parameter in Tick — Purpose and Responsibility

**Question:** What's the intent of passing `ctx` to `Tick()`? "Stop when cancelled" or "report your status when cancelled"?

**Follow-up:** Should ctx cancellation be checked at the framework level (composite nodes) or leaf level (Action/Condition)?

**Decision:** ctx is passed through for user code to use. Framework does NOT enforce cancellation checks.

**Rationale:**
- If the framework force-checks ctx before calling `fn()`, the user loses the chance to run cleanup logic inside their action.
- Composite nodes don't need to check — if a leaf returns Failure due to ctx cancellation, the composite naturally propagates that up.
- ActionFunc documentation states the contract: user is responsible for respecting ctx cancellation.

---

## 4. Fallback Re-execution Safety

**Question:** When Fallback returns Success/Failure and gets re-ticked, is `f.current = 0` safe? Could it cause issues?

**Analysis:**
- On completion (Success/Failure), `current` resets to 0 → next tick starts from child 0. This is correct.
- Child nodes that returned Success/Failure also reset their own `current` to 0. No stale state leaks.

**Related concern — Reactive vs Standard:**
- Standard Sequence (current implementation) resumes from the Running child, skipping earlier nodes. This means condition changes before the Running child are not re-evaluated.
- Reactive Sequence would re-evaluate from child 0 every tick, but requires a Halt mechanism to interrupt Running children.

**Decision:** Standard (non-reactive) for now. Reactive + Halt deferred to Phase 2.

---

## 5. Node Progress State Tracking

**Question:** Nodes only track "which child we're on" (current index), not internal progress. Is that enough?

**Decision:** Yes. This is by design in BT.

**Rationale:** BT keeps minimal state. The framework knows a node is Running and will re-tick it. Internal progress (e.g., "processed 47 of 100 files") is the responsibility of the user's Action function, not the tree. The Blackboard (Phase 2) will provide a shared data store for nodes that need to communicate progress.

---

## 6. Thread Safety / Mutex on Nodes

**Question:** No mutex on `lastStatus` or `current`. Is that safe?

**Decision:** No mutex needed. Single-goroutine tick is the contract.

**Rationale:**
- BT tick is synchronous — one goroutine walks the tree per tick.
- Adding mutex to every node adds overhead for the majority case (single-threaded).
- Concurrency control will be handled at the Tree level when the tick runner (Phase 2) is implemented, not per-node.

---

## 7. Package Structure — Flat vs Subpackages

**Question:** Should the library use subpackages (composite/, leaf/) or flat structure?

**Options considered:**
- A. Flat — everything in root package `bt`
- B. Subpackages — `composite.NewSequence`, `leaf.NewAction`
- C. Internal subpackages with root re-export — organized internally, `bt.NewSequence` externally

**Decision:** A (flat), with logical grouping by filename (composite.go, leaf.go).

**Rationale:** For a library, package boundary = API boundary. Splitting into subpackages forces users to import multiple packages for basic usage. Go BT/tree libraries are typically flat. File-level organization (composite.go, leaf.go) provides sufficient structure without import overhead.

---

## 8. Message Queue vs gRPC for Job-Dispatcher

**Question:** Should agent communication use a message queue (NATS, RabbitMQ) or gRPC?

**Concern:** User preferred a broker-mediated approach over direct connections.

**Decision:** gRPC first, with Transport interface for future MQ support.

**Rationale:**
- Push model requires dispatcher to know agent state (attributes, health) — this is inherently a strong coupling that MQ's decoupling doesn't benefit.
- With MQ, dispatcher can't intelligently route to a specific agent without per-agent queues, which defeats the purpose of smart dispatching.
- However, for unreliable networks (robots), MQ buffering has value → Transport interface allows adding MQ later.

**Follow-up — kubelet vs robot agents:**
- Stable network (kubelet-like): gRPC is natural
- Unstable network (robots): MQ's buffering shines
- Generic library → abstract Transport interface, swap implementations per environment.

---

## 9. Multi-Dispatcher Scalability

**Question:** Can multiple dispatcher instances run concurrently?

**Analysis:** Initial design had single-dispatcher assumptions (agent state in memory, SQLite).

**Decision:** Design for future multi-dispatcher from the start:
- Agent state → Redis (shared across instances, TTL for heartbeat expiry)
- Job assignment → DB-level locking (prevents duplicate assignment regardless of dispatcher count)
- JobStore and AgentStore as separate interfaces

**Current scope:** Single dispatcher. Multi-dispatcher is a future enhancement (#15).

---

## 10. nodeType() Concrete Type Dependency

**Concern (from review):** `nodeType()` in visualize.go uses type switch on concrete types. Custom Node implementations will show as "Node".

**Decision:** Acceptable for now. Will need revisiting in Phase 3 (plugin system) — possibly via a `NodeTyper` interface or node registration.
