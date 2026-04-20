package arbor

import (
	"fmt"
	"io"
	"strings"
)

func statusSymbol(s *Status) string {
	if s == nil {
		return "[ ]"
	}
	switch *s {
	case Success:
		return "[✓]"
	case Failure:
		return "[✗]"
	case Running:
		return "[~]"
	default:
		return "[?]"
	}
}

func statusLabel(s *Status) string {
	if s == nil {
		return "-"
	}
	return s.String()
}

func nodeType(n Node) string {
	switch n.(type) {
	case *Sequence:
		return "Sequence"
	case *Fallback:
		return "Fallback"
	case *Parallel:
		return "Parallel"
	case *ReactiveSequence:
		return "ReactiveSequence"
	case *ReactiveFallback:
		return "ReactiveFallback"
	case *Inverter:
		return "Inverter"
	case *Repeater:
		return "Repeater"
	case *Retry:
		return "Retry"
	case *Timeout:
		return "Timeout"
	case *Action:
		return "Action"
	case *Condition:
		return "Condition"
	default:
		return "Node"
	}
}

// PrintTree writes a visual representation of the tree to the given writer.
// Each node shows its type, name, and last tick status.
func PrintTree(w io.Writer, tree *Tree) {
	printNode(w, tree.Root(), "", true, true)
}

func printNode(w io.Writer, node Node, prefix string, isLast bool, isRoot bool) {
	var connector string
	if isRoot {
		connector = ""
	} else if isLast {
		connector = "└── "
	} else {
		connector = "├── "
	}

	var status *Status
	if s, ok := node.(Stateful); ok {
		status = s.LastStatus()
	}

	line := fmt.Sprintf("%s%s%s %s: %s (%s)\n",
		prefix, connector, statusSymbol(status),
		nodeType(node), node.String(), statusLabel(status),
	)
	fmt.Fprint(w, line)

	if p, ok := node.(Parent); ok {
		children := p.Children()
		childPrefix := prefix
		if !isRoot {
			if isLast {
				childPrefix = prefix + "    "
			} else {
				childPrefix = prefix + "│   "
			}
		}

		for i, child := range children {
			printNode(w, child, childPrefix, i == len(children)-1, false)
		}
	}
}

// SprintTree returns a string representation of the tree.
func SprintTree(tree *Tree) string {
	var sb strings.Builder
	PrintTree(&sb, tree)
	return sb.String()
}
