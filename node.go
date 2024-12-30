package restrum

import "fmt"

// node represents a single node in the routing tree.
type node struct {
	pattern  string  // the route pattern to match, e.g., /p/:lang
	part     string  // a part of the route, e.g., :lang
	children []*node // child nodes, e.g., [doc, tutorial, intro]
	isWild   bool    // whether the part contains a wildcard, e.g., :lang or *
}

// String returns a string representation of the node.
func (n *node) String() string {
	return fmt.Sprintf("node{pattern=%s, part=%s, wild=%t}", n.pattern, n.part, n.isWild)
}

// insert adds a new route pattern to the node.
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChildren(part)

	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

// search looks for a node that matches the given parts.
func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || n.isWild {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	for _, child := range n.children {
		if child.part == part || child.isWild {
			if result := child.search(parts, height+1); result != nil {
				return result
			}
		}
	}
	return nil
}

// travel collects all nodes with a non-empty pattern.
func (n *node) travel(list *[]*node) {
	if n.pattern != "" {
		*list = append(*list, n)
	}
	for _, child := range n.children {
		child.travel(list)
	}
}

// matchChildren finds a child node that matches the given part.
func (n *node) matchChildren(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// parsePattern splits a pattern into parts.
func parsePattern(pattern string) []string {
	var parts []string
	start := 0
	isWild := false

	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '/' {
			if start != i {
				parts = append(parts, pattern[start:i])
			}
			start = i + 1
		} else if pattern[i] == '*' {
			if start != i {
				parts = append(parts, pattern[start:i])
			}
			parts = append(parts, pattern[i:])
			isWild = true
			break
		}
	}

	if !isWild && start < len(pattern) {
		parts = append(parts, pattern[start:])
	}
	return parts
}
