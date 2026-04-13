package bt

import "context"

// Sequence ticks children left-to-right.
// Returns Failure immediately on first child failure.
// Returns Success when all children succeed.
// Returns Running if a child is Running (resumes from that child on next tick).
type Sequence struct {
	name     string
	children []Node
	current  int
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
			return Running
		case Failure:
			s.current = 0
			return Failure
		}
	}
	s.current = 0
	return Success
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
	name     string
	children []Node
	current  int
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
			return Running
		case Success:
			f.current = 0
			return Success
		}
	}
	f.current = 0
	return Failure
}

// Children returns the child nodes of this fallback (implements Parent).
func (f *Fallback) Children() []Node {
	return f.children
}

// String returns the name of the fallback (implements fmt.Stringer).
func (f *Fallback) String() string {
	return f.name
}
