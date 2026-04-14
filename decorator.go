package arbor

import (
	"context"
	"time"
)

// Inverter flips the child's result: Success ↔ Failure. Running passes through.
type Inverter struct {
	name       string
	child      Node
	lastStatus *Status
}

// NewInverter creates a new Inverter node with the given name and child.
func NewInverter(name string, child Node) *Inverter {
	return &Inverter{name: name, child: child}
}

// Tick executes the child node and inverts its result. Running is not inverted.
func (inv *Inverter) Tick(ctx context.Context) Status {
	status := inv.child.Tick(ctx)
	switch status {
	case Success:
		inv.lastStatus = statusPtr(Failure)
		return Failure
	case Failure:
		inv.lastStatus = statusPtr(Success)
		return Success
	default:
		inv.lastStatus = statusPtr(Running)
		return Running
	}
}

// Children returns the child node of the Inverter.
func (inv *Inverter) Children() []Node { return []Node{inv.child} }

// String returns the name of the Inverter node.
func (inv *Inverter) String() string { return inv.name }

// LastStatus returns the last status of the Inverter node.
func (inv *Inverter) LastStatus() *Status { return inv.lastStatus }

// Repeater ticks its child N times. Succeeds when all N ticks succeed.
// Fails immediately if the child fails.
type Repeater struct {
	name       string
	child      Node
	maxCount   int
	current    int
	lastStatus *Status
}

// NewRepeater creates a new Repeater node with the given name, repeater count,
// and child node.
func NewRepeater(name string, n int, child Node) *Repeater {
	return &Repeater{name: name, child: child, maxCount: n}
}

// Tick executes the child node and counts successful ticks. If the child fails,
// the count resets. The Repeater succeeds when the count reaches maxCount.
func (r *Repeater) Tick(ctx context.Context) Status {
	status := r.child.Tick(ctx)
	switch status {
	case Running:
		r.lastStatus = statusPtr(Running)
		return Running
	case Failure:
		r.current = 0
		r.lastStatus = statusPtr(Failure)
		return Failure
	case Success:
		r.current++
		if r.current >= r.maxCount {
			r.current = 0
			r.lastStatus = statusPtr(Success)
			return Success
		}
		r.lastStatus = statusPtr(Running)
		return Running
	}
	return Failure
}

// Children returns the child node of the Repeater.
func (r *Repeater) Children() []Node { return []Node{r.child} }

// String returns the name of the Repeater node.
func (r *Repeater) String() string { return r.name }

// LastStatus returns the last status of the Repeater node.
func (r *Repeater) LastStatus() *Status { return r.lastStatus }

// Retry re-ticks its child on failure, up to N attempts.
// Succeeds immediately if the child succeeds.
type Retry struct {
	name       string
	child      Node
	maxRetries int
	attempts   int
	lastStatus *Status
}

// NewRetry creates a new Retry node with the given name, max retries, and child node.
func NewRetry(name string, maxRetries int, child Node) *Retry {
	return &Retry{name: name, child: child, maxRetries: maxRetries}
}

// Tick executes the child node. On failure, increments the attempt counter
// and returns Running to retry on the next tick. Fails when max retries exhausted.
func (r *Retry) Tick(ctx context.Context) Status {
	status := r.child.Tick(ctx)
	switch status {
	case Running:
		r.lastStatus = statusPtr(Running)
		return Running
	case Success:
		r.attempts = 0
		r.lastStatus = statusPtr(Success)
		return Success
	case Failure:
		r.attempts++
		if r.attempts >= r.maxRetries {
			r.attempts = 0
			r.lastStatus = statusPtr(Failure)
			return Failure
		}
		r.lastStatus = statusPtr(Running)
		return Running
	}
	return Failure
}

// Children returns the child node of the Retry.
func (r *Retry) Children() []Node { return []Node{r.child} }

// String returns the name of the Retry node.
func (r *Retry) String() string { return r.name }

// LastStatus returns the last status of the Retry node.
func (r *Retry) LastStatus() *Status { return r.lastStatus }

// Timeout fails the child if it stays Running beyond the given duration.
// Tracks elapsed time across ticks.
type Timeout struct {
	name       string
	child      Node
	duration   time.Duration
	startTime  time.Time
	running    bool
	lastStatus *Status
}

// NewTimeout creates a new Timeout node with the given name, duration, and child node.
func NewTimeout(name string, duration time.Duration, child Node) *Timeout {
	return &Timeout{name: name, child: child, duration: duration}
}

// Tick executes the child node. If the child returns Running, starts or continues
// tracking elapsed time. Fails if the duration is exceeded.
func (t *Timeout) Tick(ctx context.Context) Status {
	status := t.child.Tick(ctx)
	switch status {
	case Running:
		if !t.running {
			t.startTime = time.Now()
			t.running = true
		}
		if time.Since(t.startTime) >= t.duration {
			t.running = false
			t.lastStatus = statusPtr(Failure)
			return Failure
		}
		t.lastStatus = statusPtr(Running)
		return Running
	default:
		t.running = false
		t.lastStatus = statusPtr(status)
		return status
	}
}

// Children returns the child node of the Timeout.
func (t *Timeout) Children() []Node { return []Node{t.child} }

// String returns the name of the Timeout node.
func (t *Timeout) String() string { return t.name }

// LastStatus returns the last status of the Timeout node.
func (t *Timeout) LastStatus() *Status { return t.lastStatus }
