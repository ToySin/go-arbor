package arbor

import (
	"context"
	"errors"
	"time"
)

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

// TickEvent contains information about a completed tick.
type TickEvent struct {
	Tick   int
	Status Status
}

// RunOption configures the behavior of Tree.Run.
type RunOption func(*runConfig)

type runConfig struct {
	callback func(TickEvent) bool
}

// WithTickCallback registers a function called after each tick.
// Return true to continue, false to stop the run loop early.
func WithTickCallback(fn func(TickEvent) bool) RunOption {
	return func(cfg *runConfig) {
		cfg.callback = fn
	}
}

// Run executes the tree's tick loop at the given interval until the context
// is cancelled or a callback returns false.
// Returns the context's error when the loop exits due to cancellation,
// or nil when a callback stops the loop.
func (t *Tree) Run(ctx context.Context, interval time.Duration, opts ...RunOption) error {
	if interval <= 0 {
		return errors.New("arbor: interval must be positive")
	}
	cfg := &runConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	tick := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			tick++
			status := t.Tick(ctx)
			if cfg.callback != nil {
				if !cfg.callback(TickEvent{Tick: tick, Status: status}) {
					return nil
				}
			}
		}
	}
}
