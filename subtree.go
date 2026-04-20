package arbor

import "context"

// SubtreeOption configures a Subtree node.
type SubtreeOption func(*Subtree)

// WithInputMapping maps a key from the parent Blackboard to the subtree Blackboard.
// Before each tick, the value is copied from parentKey to subtreeKey.
func WithInputMapping(parentKey, subtreeKey string) SubtreeOption {
	return func(s *Subtree) {
		s.inputMappings = append(s.inputMappings, keyMapping{parentKey, subtreeKey})
	}
}

// WithOutputMapping maps a key from the subtree Blackboard to the parent Blackboard.
// After each tick, the value is copied from subtreeKey to parentKey.
func WithOutputMapping(subtreeKey, parentKey string) SubtreeOption {
	return func(s *Subtree) {
		s.outputMappings = append(s.outputMappings, keyMapping{subtreeKey, parentKey})
	}
}

type keyMapping struct {
	from string
	to   string
}

// Subtree wraps a Tree as a node, enabling modular tree composition.
// The inner tree has its own isolated Blackboard.
// Data can be exchanged via input/output key mappings.
type Subtree struct {
	name           string
	inner          *Tree
	inputMappings  []keyMapping
	outputMappings []keyMapping
	lastStatus     *Status
}

// NewSubtree creates a new Subtree node wrapping the given tree.
func NewSubtree(name string, inner *Tree, opts ...SubtreeOption) *Subtree {
	s := &Subtree{
		name:  name,
		inner: inner,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Tick copies input mappings, ticks the inner tree, then copies output mappings.
func (s *Subtree) Tick(ctx context.Context) Status {
	parentBB := BlackboardFrom(ctx)
	innerBB := s.inner.Blackboard()

	// Copy inputs: parent → subtree
	if parentBB != nil {
		for _, m := range s.inputMappings {
			if v, ok := parentBB.Get(m.from); ok {
				innerBB.Set(m.to, v)
			}
		}
	}

	status := s.inner.Tick(ctx)

	// Copy outputs: subtree → parent
	if parentBB != nil {
		for _, m := range s.outputMappings {
			if v, ok := innerBB.Get(m.from); ok {
				parentBB.Set(m.to, v)
			}
		}
	}

	s.lastStatus = statusPtr(status)
	return status
}

// Halt propagates halt to the inner tree's root node.
func (s *Subtree) Halt() {
	haltNode(s.inner.Root())
	s.lastStatus = nil
}

// Children returns the inner tree's root node for visualization.
func (s *Subtree) Children() []Node {
	return []Node{s.inner.Root()}
}

// String returns the name of the subtree.
func (s *Subtree) String() string {
	return s.name
}

// LastStatus returns the last tick result (implements Stateful).
func (s *Subtree) LastStatus() *Status {
	return s.lastStatus
}
