package bt

import "context"

// ActionFunc is a function that performs work and returns a status.
//
// The function should not block for long periods of time;
// if it needs to perform long-running work, it should return Running
// and continue the work on subsequent ticks.
// The function should be thread-safe and reentrant, as it may be called
// concurrently from multiple ticks.
//
// The function should not modify the context, as it may be shared across ticks.
// The function should not panic; if it encounters an error,
// it should return Failure. The function should be terminated by the caller
// using the context's cancellation mechanism.
type ActionFunc func(ctx context.Context) Status

// Action is a leaf node that executes a user-provided function.
type Action struct {
	name string
	fn   ActionFunc
}

// NewAction creates a new Action node with the given name and function.
func NewAction(name string, fn ActionFunc) *Action {
	return &Action{
		name: name,
		fn:   fn,
	}
}

// Tick executes the action's function and returns its status.
func (a *Action) Tick(ctx context.Context) Status {
	return a.fn(ctx)
}

// String returns the name of the action (implements fmt.Stringer).
func (a *Action) String() string {
	return a.name
}

// ConditionFunc is a predicate that evaluates to true or false.
type ConditionFunc func(ctx context.Context) bool

// Condition is a leaf node that evaluates a predicate.
// Returns Success if true, Failure if false. Never returns Running.
type Condition struct {
	name string
	fn   ConditionFunc
}

// NewCondition creates a new Condition node with the given name
// and predicate function.
func NewCondition(name string, fn ConditionFunc) *Condition {
	return &Condition{
		name: name,
		fn:   fn,
	}
}

// Tick evaluates the condition's predicate and returns Success if true,
// Failure if false. Never returns Running.
func (c *Condition) Tick(ctx context.Context) Status {
	if c.fn(ctx) {
		return Success
	}
	return Failure
}

// String returns the name of the condition (implements fmt.Stringer).
func (c *Condition) String() string {
	return c.name
}
