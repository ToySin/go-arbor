package arbor

import "context"

// ReactiveSequence ticks children from the first child on every tick.
// Unlike Sequence, it does NOT resume from the previously Running child.
// If a child that was previously Success now fails, Running children are halted.
// This enables "guard condition" patterns where conditions are re-checked every tick.
type ReactiveSequence struct {
	name       string
	children   []Node
	running    int // index of the currently Running child, -1 if none
	lastStatus *Status
}

// NewReactiveSequence creates a new ReactiveSequence node.
func NewReactiveSequence(name string, children ...Node) *ReactiveSequence {
	return &ReactiveSequence{
		name:     name,
		children: children,
		running:  -1,
	}
}

// Tick always starts from child 0. If a previously Running child is no longer
// the active Running node, it is halted.
func (rs *ReactiveSequence) Tick(ctx context.Context) Status {
	for i, child := range rs.children {
		status := child.Tick(ctx)
		switch status {
		case Running:
			// If a different child was Running before, halt it
			if rs.running != -1 && rs.running != i {
				haltNode(rs.children[rs.running])
			}
			rs.running = i
			rs.lastStatus = statusPtr(Running)
			return Running
		case Failure:
			// Halt the previously Running child if any
			if rs.running != -1 {
				haltNode(rs.children[rs.running])
				rs.running = -1
			}
			rs.lastStatus = statusPtr(Failure)
			return Failure
		}
	}
	// All succeeded
	rs.running = -1
	rs.lastStatus = statusPtr(Success)
	return Success
}

// Halt interrupts the ReactiveSequence and halts the Running child.
func (rs *ReactiveSequence) Halt() {
	if rs.running != -1 {
		haltNode(rs.children[rs.running])
		rs.running = -1
	}
	rs.lastStatus = nil
}

// Children returns the child nodes (implements Parent).
func (rs *ReactiveSequence) Children() []Node { return rs.children }

// String returns the name (implements fmt.Stringer).
func (rs *ReactiveSequence) String() string { return rs.name }

// LastStatus returns the last tick result (implements Stateful).
func (rs *ReactiveSequence) LastStatus() *Status { return rs.lastStatus }

// ReactiveFallback ticks children from the first child on every tick.
// Unlike Fallback, it does NOT resume from the previously Running child.
// If a higher-priority child now succeeds, the previously Running child is halted.
type ReactiveFallback struct {
	name       string
	children   []Node
	running    int // index of the currently Running child, -1 if none
	lastStatus *Status
}

// NewReactiveFallback creates a new ReactiveFallback node.
func NewReactiveFallback(name string, children ...Node) *ReactiveFallback {
	return &ReactiveFallback{
		name:     name,
		children: children,
		running:  -1,
	}
}

// Tick always starts from child 0. If a higher-priority child succeeds,
// the previously Running lower-priority child is halted.
func (rf *ReactiveFallback) Tick(ctx context.Context) Status {
	for i, child := range rf.children {
		status := child.Tick(ctx)
		switch status {
		case Running:
			// If a different child was Running before, halt it
			if rf.running != -1 && rf.running != i {
				haltNode(rf.children[rf.running])
			}
			rf.running = i
			rf.lastStatus = statusPtr(Running)
			return Running
		case Success:
			// Halt the previously Running child if any
			if rf.running != -1 {
				haltNode(rf.children[rf.running])
				rf.running = -1
			}
			rf.lastStatus = statusPtr(Success)
			return Success
		}
	}
	// All failed
	rf.running = -1
	rf.lastStatus = statusPtr(Failure)
	return Failure
}

// Halt interrupts the ReactiveFallback and halts the Running child.
func (rf *ReactiveFallback) Halt() {
	if rf.running != -1 {
		haltNode(rf.children[rf.running])
		rf.running = -1
	}
	rf.lastStatus = nil
}

// Children returns the child nodes (implements Parent).
func (rf *ReactiveFallback) Children() []Node { return rf.children }

// String returns the name (implements fmt.Stringer).
func (rf *ReactiveFallback) String() string { return rf.name }

// LastStatus returns the last tick result (implements Stateful).
func (rf *ReactiveFallback) LastStatus() *Status { return rf.lastStatus }
