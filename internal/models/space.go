// Package models defines the data structures for Spacelift resources.
package models

// Space represents a Spacelift space.
type Space struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	ParentSpace    *string  `json:"parentSpace,omitempty"`
	InheritEntities bool    `json:"inheritEntities"`
	Labels         []string `json:"labels"`
}

// SpaceTree represents a space with its children for hierarchical display.
type SpaceTree struct {
	Space    Space        `json:"space"`
	Children []*SpaceTree `json:"children,omitempty"`
}

// BuildSpaceTree builds a hierarchical tree from a flat list of spaces.
func BuildSpaceTree(spaces []Space) []*SpaceTree {
	// Create a map of space ID to SpaceTree node
	nodeMap := make(map[string]*SpaceTree)
	for _, s := range spaces {
		space := s // Create a copy to avoid pointer issues
		nodeMap[s.ID] = &SpaceTree{Space: space}
	}

	// Build the tree by linking children to parents
	var roots []*SpaceTree
	for _, node := range nodeMap {
		if node.Space.ParentSpace == nil || *node.Space.ParentSpace == "" {
			// Root node
			roots = append(roots, node)
		} else {
			// Find parent and add as child
			if parent, ok := nodeMap[*node.Space.ParentSpace]; ok {
				parent.Children = append(parent.Children, node)
			} else {
				// Parent not found, treat as root
				roots = append(roots, node)
			}
		}
	}

	return roots
}

// FlattenSpaceTree flattens a space tree back to a list (depth-first).
func FlattenSpaceTree(trees []*SpaceTree) []Space {
	var result []Space
	var flatten func(node *SpaceTree, depth int)
	flatten = func(node *SpaceTree, depth int) {
		result = append(result, node.Space)
		for _, child := range node.Children {
			flatten(child, depth+1)
		}
	}

	for _, tree := range trees {
		flatten(tree, 0)
	}

	return result
}
