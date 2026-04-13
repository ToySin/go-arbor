package arbor

import "context"

// Tree is the top-level container that drives tick execution from the root node.
type Tree struct {
	root Node
}

// NewTree creates a new behavior tree with the given root node.
func NewTree(root Node) *Tree {
	return &Tree{root: root}
}

// Tick executes a single tick starting from the root node.
func (t *Tree) Tick(ctx context.Context) Status {
	return t.root.Tick(ctx)
}

// Root returns the root node of the tree.
func (t *Tree) Root() Node {
	return t.root
}
