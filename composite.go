package arbor

import "context"

// Sequence ticks children left-to-right.
// Returns Failure immediately on first child failure.
// Returns Success when all children succeed.
// Returns Running if a child is Running (resumes from that child on next tick).
type Sequence struct {
	name       string
	children   []Node
	current    int
	lastStatus *Status
}

// NewSequence creates a new Sequence node with the given name and child nodes.
func NewSequence(name string, children ...Node) *Sequence {
	return &Sequence{
		name:     name,
		children: children,
	}
}

// Tick executes the sequence's children in order and returns
// the appropriate status.
func (s *Sequence) Tick(ctx context.Context) Status {
	for i := s.current; i < len(s.children); i++ {
		status := s.children[i].Tick(ctx)
		switch status {
		case Running:
			s.current = i
			s.lastStatus = statusPtr(status)
			return Running
		case Failure:
			s.current = 0
			s.lastStatus = statusPtr(status)
			return Failure
		}
	}
	s.current = 0
	s.lastStatus = statusPtr(Success)
	return Success
}

// Halt interrupts the Sequence and halts the currently Running child.
func (s *Sequence) Halt() {
	if s.current < len(s.children) {
		haltNode(s.children[s.current])
	}
	s.current = 0
	s.lastStatus = nil
}

// LastStatus returns the result of the most recent tick (implements Stateful).
func (s *Sequence) LastStatus() *Status {
	return s.lastStatus
}

// Children returns the child nodes of this sequence (implements Parent).
func (s *Sequence) Children() []Node {
	return s.children
}

// String returns the name of the sequence (implements fmt.Stringer).
func (s *Sequence) String() string {
	return s.name
}

// Fallback (Selector) ticks children left-to-right.
// Returns Success immediately on first child success.
// Returns Failure when all children fail.
// Returns Running if a child is Running (resumes from that child on next tick).
type Fallback struct {
	name       string
	children   []Node
	current    int
	lastStatus *Status
}

// NewFallback creates a new Fallback node with the given name and child nodes.
func NewFallback(name string, children ...Node) *Fallback {
	return &Fallback{
		name:     name,
		children: children,
	}
}

// Tick executes the fallback's children in order and returns
// the appropriate status.
func (f *Fallback) Tick(ctx context.Context) Status {
	for i := f.current; i < len(f.children); i++ {
		status := f.children[i].Tick(ctx)
		switch status {
		case Running:
			f.current = i
			f.lastStatus = statusPtr(status)
			return Running
		case Success:
			f.current = 0
			f.lastStatus = statusPtr(status)
			return Success
		}
	}
	f.current = 0
	f.lastStatus = statusPtr(Failure)
	return Failure
}

// Halt interrupts the Fallback and halts the currently Running child.
func (f *Fallback) Halt() {
	if f.current < len(f.children) {
		haltNode(f.children[f.current])
	}
	f.current = 0
	f.lastStatus = nil
}

// LastStatus returns the result of the most recent tick (implements Stateful).
func (f *Fallback) LastStatus() *Status {
	return f.lastStatus
}

// Children returns the child nodes of this fallback (implements Parent).
func (f *Fallback) Children() []Node {
	return f.children
}

// String returns the name of the fallback (implements fmt.Stringer).
func (f *Fallback) String() string {
	return f.name
}

// ParallelPolicy defines when a Parallel node should return Success or Failure.
// ParallelOption configures a Parallel node.
type ParallelOption func(*Parallel)

// WithSuccessThreshold sets the number of children that must succeed
// for the Parallel node to return Success.
func WithSuccessThreshold(n int) ParallelOption {
	return func(p *Parallel) {
		p.successThreshold = n
	}
}

// WithFailureThreshold sets the number of children that must fail
// for the Parallel node to return Failure.
func WithFailureThreshold(n int) ParallelOption {
	return func(p *Parallel) {
		p.failureThreshold = n
	}
}

// Parallel ticks all non-completed children on each tick.
// By default, all children must succeed (successThreshold = len(children))
// and one failure is enough to fail (failureThreshold = 1).
type Parallel struct {
	name             string
	children         []Node
	successThreshold int
	failureThreshold int
	completed        []Status
	done             []bool
	lastStatus       *Status
}

// NewParallel creates a new Parallel node with the given name and child nodes.
// Default policy: all children must succeed, one failure causes Failure.
func NewParallel(name string, children []Node, opts ...ParallelOption) *Parallel {
	p := &Parallel{
		name:             name,
		children:         children,
		successThreshold: len(children),
		failureThreshold: 1,
		completed:        make([]Status, len(children)),
		done:             make([]bool, len(children)),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Tick executes all non-completed children and evaluates the policy thresholds.
func (p *Parallel) Tick(ctx context.Context) Status {
	successCount := 0
	failureCount := 0

	for i, child := range p.children {
		if p.done[i] {
			if p.completed[i] == Success {
				successCount++
			} else {
				failureCount++
			}
			continue
		}

		status := child.Tick(ctx)
		switch status {
		case Success:
			p.done[i] = true
			p.completed[i] = Success
			successCount++
		case Failure:
			p.done[i] = true
			p.completed[i] = Failure
			failureCount++
		}

		if successCount >= p.successThreshold {
			p.reset()
			p.lastStatus = statusPtr(Success)
			return Success
		}
		if failureCount >= p.failureThreshold {
			p.reset()
			p.lastStatus = statusPtr(Failure)
			return Failure
		}
	}

	p.lastStatus = statusPtr(Running)
	return Running
}

func (p *Parallel) reset() {
	for i := range p.done {
		p.done[i] = false
	}
}

// Halt interrupts the Parallel and halts all Running children.
func (p *Parallel) Halt() {
	for i, child := range p.children {
		if !p.done[i] {
			haltNode(child)
		}
	}
	p.reset()
	p.lastStatus = nil
}

// LastStatus returns the result of the most recent tick (implements Stateful).
func (p *Parallel) LastStatus() *Status {
	return p.lastStatus
}

// Children returns the child nodes of this parallel (implements Parent).
func (p *Parallel) Children() []Node {
	return p.children
}

// String returns the name of the parallel (implements fmt.Stringer).
func (p *Parallel) String() string {
	return p.name
}
