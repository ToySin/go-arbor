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

func NewInverter(name string, child Node) *Inverter {
	return &Inverter{name: name, child: child}
}

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

func (inv *Inverter) Children() []Node   { return []Node{inv.child} }
func (inv *Inverter) String() string     { return inv.name }
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

func NewRepeater(name string, n int, child Node) *Repeater {
	return &Repeater{name: name, child: child, maxCount: n}
}

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

func (r *Repeater) Children() []Node   { return []Node{r.child} }
func (r *Repeater) String() string     { return r.name }
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

func NewRetry(name string, maxRetries int, child Node) *Retry {
	return &Retry{name: name, child: child, maxRetries: maxRetries}
}

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

func (r *Retry) Children() []Node   { return []Node{r.child} }
func (r *Retry) String() string     { return r.name }
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

func NewTimeout(name string, duration time.Duration, child Node) *Timeout {
	return &Timeout{name: name, child: child, duration: duration}
}

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

func (t *Timeout) Children() []Node   { return []Node{t.child} }
func (t *Timeout) String() string     { return t.name }
func (t *Timeout) LastStatus() *Status { return t.lastStatus }
