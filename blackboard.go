package arbor

import "context"

type contextKey string

const blackboardKey contextKey = "arbor.blackboard"

// Blackboard is a shared key-value store for passing data between nodes.
type Blackboard struct {
	data map[string]any
}

// NewBlackboard creates a new empty Blackboard.
func NewBlackboard() *Blackboard {
	return &Blackboard{data: make(map[string]any)}
}

// Set stores a value under the given key.
func (bb *Blackboard) Set(key string, value any) {
	bb.data[key] = value
}

// Get retrieves a value by key. Returns the value and whether the key exists.
func (bb *Blackboard) Get(key string) (any, bool) {
	v, ok := bb.data[key]
	return v, ok
}

// Delete removes a key from the blackboard.
func (bb *Blackboard) Delete(key string) {
	delete(bb.data, key)
}

// Has returns whether the given key exists.
func (bb *Blackboard) Has(key string) bool {
	_, ok := bb.data[key]
	return ok
}

// Clear removes all entries from the blackboard.
func (bb *Blackboard) Clear() {
	clear(bb.data)
}

// GetTyped retrieves a value by key with type assertion.
// Returns the zero value and false if the key doesn't exist or the type doesn't match.
func GetTyped[T any](bb *Blackboard, key string) (T, bool) {
	v, ok := bb.data[key]
	if !ok {
		var zero T
		return zero, false
	}
	typed, ok := v.(T)
	if !ok {
		var zero T
		return zero, false
	}
	return typed, true
}

// WithBlackboard returns a new context with the given Blackboard attached.
func WithBlackboard(ctx context.Context, bb *Blackboard) context.Context {
	return context.WithValue(ctx, blackboardKey, bb)
}

// BlackboardFrom retrieves the Blackboard from the context.
// Returns nil if no Blackboard is attached.
func BlackboardFrom(ctx context.Context) *Blackboard {
	bb, _ := ctx.Value(blackboardKey).(*Blackboard)
	return bb
}
