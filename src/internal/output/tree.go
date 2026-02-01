//nolint:errcheck // CLI output errors are unrecoverable
package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/randlee/claude-history/pkg/agent"
)

// WriteTree writes an agent tree in the specified format.
func WriteTree(w io.Writer, tree *agent.TreeNode, format Format) error {
	switch format {
	case FormatJSON:
		return WriteJSON(w, tree)
	case FormatDOT:
		return writeTreeDOT(w, tree)
	default:
		return writeTreeASCII(w, tree)
	}
}

func writeTreeASCII(w io.Writer, tree *agent.TreeNode) error {
	if tree == nil {
		return nil
	}

	// Write root (main session)
	fmt.Fprintf(w, "Session: %s\n", tree.SessionID)
	fmt.Fprintf(w, "├── Main conversation (%d entries)\n", tree.EntryCount)

	// Write children
	for i, child := range tree.Children {
		isLast := i == len(tree.Children)-1
		writeTreeNodeASCII(w, child, "", isLast)
	}

	return nil
}

func writeTreeNodeASCII(w io.Writer, node *agent.TreeNode, prefix string, isLast bool) {
	if node == nil {
		return
	}

	// Determine the connector
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	// Build the label
	label := node.AgentID
	if node.AgentType != "" {
		label = fmt.Sprintf("%s (%s)", label, node.AgentType)
	}

	// Write the node
	fmt.Fprintf(w, "%s%s%s\n", prefix, connector, label)

	// Write entry count on next line
	childPrefix := prefix
	if isLast {
		childPrefix += "    "
	} else {
		childPrefix += "│   "
	}
	fmt.Fprintf(w, "%s└── %d entries\n", childPrefix, node.EntryCount)

	// Write children (if any nested agents)
	for i, child := range node.Children {
		childIsLast := i == len(node.Children)-1
		writeTreeNodeASCII(w, child, childPrefix, childIsLast)
	}
}

func writeTreeDOT(w io.Writer, tree *agent.TreeNode) error {
	if tree == nil {
		return nil
	}

	fmt.Fprintln(w, "digraph AgentTree {")
	fmt.Fprintln(w, "  rankdir=TB;")
	fmt.Fprintln(w, "  node [shape=box];")
	fmt.Fprintln(w)

	// Write root node
	rootLabel := fmt.Sprintf("Session\\n%s\\n(%d entries)", tree.SessionID, tree.EntryCount)
	rootID := sanitizeDOTID(tree.SessionID)
	fmt.Fprintf(w, "  %s [label=\"%s\"];\n", rootID, rootLabel)

	// Write child nodes
	for _, child := range tree.Children {
		writeTreeNodeDOT(w, child, rootID)
	}

	fmt.Fprintln(w, "}")
	return nil
}

func writeTreeNodeDOT(w io.Writer, node *agent.TreeNode, parentID string) {
	if node == nil {
		return
	}

	nodeID := sanitizeDOTID(node.AgentID)
	label := node.AgentID
	if node.AgentType != "" {
		label = fmt.Sprintf("%s\\n(%s)", label, node.AgentType)
	}
	label = fmt.Sprintf("%s\\n%d entries", label, node.EntryCount)

	fmt.Fprintf(w, "  %s [label=\"%s\"];\n", nodeID, label)
	fmt.Fprintf(w, "  %s -> %s;\n", parentID, nodeID)

	for _, child := range node.Children {
		writeTreeNodeDOT(w, child, nodeID)
	}
}

// sanitizeDOTID converts a string to a valid DOT identifier.
func sanitizeDOTID(s string) string {
	// Replace dashes and other special chars with underscores
	result := strings.ReplaceAll(s, "-", "_")
	result = strings.ReplaceAll(result, ".", "_")
	return result
}
