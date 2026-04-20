package arbor

import (
	"fmt"
	"time"
)

// Builder provides a fluent API for constructing behavior trees.
//
// Composite and decorator methods open a new scope that collects children
// until End() is called. Leaf methods add nodes directly to the current scope.
//
// Example:
//
//	tree, err := arbor.NewBuilder().
//	    Sequence("dispatch").
//	        Condition("agent-idle", isIdle).
//	        Action("assign-job", assignJob).
//	        Action("notify", notify).
//	    End().
//	    Build()
type Builder struct {
	stack []*builderScope
	err   error
}

type builderScope struct {
	name           string
	children       []Node
	buildComposite func(children []Node) Node
	buildDecorator func(child Node) Node
}

func (s *builderScope) isDecorator() bool {
	return s.buildDecorator != nil
}

// NewBuilder creates a new Builder with an empty root scope.
func NewBuilder() *Builder {
	return &Builder{
		stack: []*builderScope{{}},
	}
}

func (b *Builder) current() *builderScope {
	return b.stack[len(b.stack)-1]
}

func (b *Builder) pushComposite(name string, build func(children []Node) Node) *Builder {
	if b.err != nil {
		return b
	}
	b.stack = append(b.stack, &builderScope{
		name:           name,
		buildComposite: build,
	})
	return b
}

func (b *Builder) pushDecorator(name string, build func(child Node) Node) *Builder {
	if b.err != nil {
		return b
	}
	b.stack = append(b.stack, &builderScope{
		name:           name,
		buildDecorator: build,
	})
	return b
}

func (b *Builder) addLeaf(node Node) *Builder {
	if b.err != nil {
		return b
	}
	b.current().children = append(b.current().children, node)
	return b
}

// Sequence opens a Sequence scope.
func (b *Builder) Sequence(name string) *Builder {
	return b.pushComposite(name, func(children []Node) Node {
		return NewSequence(name, children...)
	})
}

// Fallback opens a Fallback scope.
func (b *Builder) Fallback(name string) *Builder {
	return b.pushComposite(name, func(children []Node) Node {
		return NewFallback(name, children...)
	})
}

// Parallel opens a Parallel scope.
func (b *Builder) Parallel(name string, opts ...ParallelOption) *Builder {
	return b.pushComposite(name, func(children []Node) Node {
		return NewParallel(name, children, opts...)
	})
}

// Inverter opens an Inverter scope (expects exactly one child before End).
func (b *Builder) Inverter(name string) *Builder {
	return b.pushDecorator(name, func(child Node) Node {
		return NewInverter(name, child)
	})
}

// Repeater opens a Repeater scope (expects exactly one child before End).
func (b *Builder) Repeater(name string, n int) *Builder {
	return b.pushDecorator(name, func(child Node) Node {
		return NewRepeater(name, n, child)
	})
}

// Retry opens a Retry scope (expects exactly one child before End).
func (b *Builder) Retry(name string, maxRetries int) *Builder {
	return b.pushDecorator(name, func(child Node) Node {
		return NewRetry(name, maxRetries, child)
	})
}

// Timeout opens a Timeout scope (expects exactly one child before End).
func (b *Builder) Timeout(name string, d time.Duration) *Builder {
	return b.pushDecorator(name, func(child Node) Node {
		return NewTimeout(name, d, child)
	})
}

// Action adds an Action leaf to the current scope.
func (b *Builder) Action(name string, fn ActionFunc) *Builder {
	return b.addLeaf(NewAction(name, fn))
}

// Condition adds a Condition leaf to the current scope.
func (b *Builder) Condition(name string, fn ConditionFunc) *Builder {
	return b.addLeaf(NewCondition(name, fn))
}

// End closes the current scope and adds the resulting node to the parent scope.
func (b *Builder) End() *Builder {
	if b.err != nil {
		return b
	}
	if len(b.stack) <= 1 {
		b.err = fmt.Errorf("arbor: End() called without matching open scope")
		return b
	}

	current := b.current()
	b.stack = b.stack[:len(b.stack)-1]

	if current.isDecorator() {
		if len(current.children) != 1 {
			b.err = fmt.Errorf("arbor: decorator %q expects exactly 1 child, got %d", current.name, len(current.children))
			return b
		}
		b.current().children = append(b.current().children, current.buildDecorator(current.children[0]))
	} else {
		if len(current.children) == 0 {
			b.err = fmt.Errorf("arbor: composite %q has no children", current.name)
			return b
		}
		b.current().children = append(b.current().children, current.buildComposite(current.children))
	}
	return b
}

// Build validates the tree structure and returns the constructed Tree.
func (b *Builder) Build() (*Tree, error) {
	if b.err != nil {
		return nil, b.err
	}
	if len(b.stack) != 1 {
		return nil, fmt.Errorf("arbor: %d unclosed scope(s), missing End() call(s)", len(b.stack)-1)
	}
	root := b.stack[0]
	if len(root.children) != 1 {
		return nil, fmt.Errorf("arbor: tree must have exactly 1 root node, got %d", len(root.children))
	}
	return NewTree(root.children[0]), nil
}

// MustBuild calls Build and panics if the tree is invalid.
func (b *Builder) MustBuild() *Tree {
	tree, err := b.Build()
	if err != nil {
		panic(err)
	}
	return tree
}
