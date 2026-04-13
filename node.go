package bt

import (
	"context"
	"fmt"
)

// Status represents the result of a single tick on a node.
type Status int

const (
	// Success indicates the node completed successfully.
	Success Status = iota
	// Failure indicates the node failed.
	Failure
	// Running indicates the node is still in progress and should be ticked again.
	Running
)

// String returns a human-readable represenation of the Status.
func (s Status) String() string {
	switch s {
	case Success:
		return "Success"
	case Failure:
		return "Failure"
	case Running:
		return "Running"
	default:
		return fmt.Sprintf("Unknown(%d)", int(s))
	}
}

// Node is the fundamental interface that all behavior tree nodes implement.
type Node interface {
	// Tick executes a single tick of this node and returns its status.
	Tick(ctx context.Context) Status

	// String returns a human-readable name for this node (implements fmt.Stringer).
	String() string
}

// Parent is implemented by nodes that have children (composite and decorator nodes).
type Parent interface {
	// Children returns the child nodes of this parent node.
	Children() []Node
}

// Stateful is implemented by nodes that track their last tick result.
// Used by the visualization system to display node status.
type Stateful interface {
	// LastStatus returns the result of the most recent tick,
	// or nil if the node has not been ticked yet.
	LastStatus() *Status
}
