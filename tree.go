package arbor

import "context"

// Tree is the top-level container that drives tick execution from the root node.
type Tree struct {
	root       Node
	blackboard *Blackboard
}

// NewTree creates a new behavior tree with the given root node.
// A Blackboard is automatically created and injected into the context on each tick.
func NewTree(root Node) *Tree {
	return &Tree{
		root:       root,
		blackboard: NewBlackboard(),
	}
}

// Tick executes a single tick starting from the root node.
// The tree's Blackboard is injected into the context automatically.
func (t *Tree) Tick(ctx context.Context) Status {
	ctx = WithBlackboard(ctx, t.blackboard)
	return t.root.Tick(ctx)
}

// Blackboard returns the tree's shared Blackboard.
func (t *Tree) Blackboard() *Blackboard {
	return t.blackboard
}

// Root returns the root node of the tree.
func (t *Tree) Root() Node {
	return t.root
}
