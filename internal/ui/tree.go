// Package ui provides terminal UI utilities for SpaceBridge.
package ui

import (
	"fmt"
	"strings"

	"github.com/jnesspace/spacebridge/internal/models"
)

// TreeStyle defines the characters used for tree rendering.
type TreeStyle struct {
	Pipe   string
	Tee    string
	Corner string
	Blank  string
}

// DefaultTreeStyle returns the default tree rendering style.
func DefaultTreeStyle() TreeStyle {
	return TreeStyle{
		Pipe:   "│   ",
		Tee:    "├── ",
		Corner: "└── ",
		Blank:  "    ",
	}
}

// RenderSpaceTree renders a space tree as a formatted string.
func RenderSpaceTree(trees []*models.SpaceTree) string {
	var sb strings.Builder
	style := DefaultTreeStyle()

	for i, tree := range trees {
		isLast := i == len(trees)-1
		renderSpaceNode(&sb, tree, "", isLast, style)
	}

	return sb.String()
}

// renderSpaceNode recursively renders a space tree node.
func renderSpaceNode(sb *strings.Builder, node *models.SpaceTree, prefix string, isLast bool, style TreeStyle) {
	// Determine the connector
	connector := style.Tee
	if isLast {
		connector = style.Corner
	}

	// For root nodes, don't add a connector
	if prefix == "" {
		sb.WriteString(formatSpaceInfo(node.Space))
	} else {
		sb.WriteString(prefix + connector + formatSpaceInfo(node.Space))
	}
	sb.WriteString("\n")

	// Calculate the prefix for children
	childPrefix := prefix
	if prefix != "" {
		if isLast {
			childPrefix += style.Blank
		} else {
			childPrefix += style.Pipe
		}
	}

	// Render children
	for i, child := range node.Children {
		childIsLast := i == len(node.Children)-1
		renderSpaceNode(sb, child, childPrefix, childIsLast, style)
	}
}

// formatSpaceInfo formats space information for display.
func formatSpaceInfo(space models.Space) string {
	info := fmt.Sprintf("%s", space.ID)
	if space.Name != space.ID && space.Name != "" {
		info += fmt.Sprintf(" (%s)", space.Name)
	}
	if len(space.Labels) > 0 {
		info += fmt.Sprintf(" [%s]", strings.Join(space.Labels, ", "))
	}
	return info
}

// RenderTable renders data in a simple table format.
func RenderTable(headers []string, rows [][]string) string {
	if len(rows) == 0 {
		return "No data to display\n"
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var sb strings.Builder

	// Render header
	sb.WriteString(renderRow(headers, widths))
	sb.WriteString(renderSeparator(widths))

	// Render rows
	for _, row := range rows {
		sb.WriteString(renderRow(row, widths))
	}

	return sb.String()
}

// renderRow renders a single table row.
func renderRow(cells []string, widths []int) string {
	var parts []string
	for i, cell := range cells {
		if i < len(widths) {
			parts = append(parts, fmt.Sprintf("%-*s", widths[i], cell))
		}
	}
	return strings.Join(parts, " │ ") + "\n"
}

// renderSeparator renders a table separator line.
func renderSeparator(widths []int) string {
	var parts []string
	for _, w := range widths {
		parts = append(parts, strings.Repeat("─", w))
	}
	return strings.Join(parts, "─┼─") + "\n"
}
