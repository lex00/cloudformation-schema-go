package codegen

import "sort"

// TopologicalSort returns nodes sorted by dependencies (dependencies first).
// Uses Kahn's algorithm for stable topological ordering.
//
// Parameters:
//   - nodes: slice of node identifiers to sort
//   - getDeps: function that returns dependencies for a given node
//
// Returns nodes in dependency order. If cycles exist, remaining nodes
// are appended at the end in sorted order.
//
// Example:
//
//	nodes := []string{"A", "B", "C"}
//	deps := map[string][]string{
//	    "A": {"B"},      // A depends on B
//	    "B": {"C"},      // B depends on C
//	    "C": {},         // C has no dependencies
//	}
//	sorted := TopologicalSort(nodes, func(n string) []string {
//	    return deps[n]
//	})
//	// Result: ["C", "B", "A"]
func TopologicalSort(nodes []string, getDeps func(string) []string) []string {
	// Build in-degree map (count of dependencies)
	inDegree := make(map[string]int)
	for _, node := range nodes {
		inDegree[node] = len(getDeps(node))
	}

	// Start with nodes that have no dependencies (in-degree 0)
	var queue []string
	for _, node := range nodes {
		if inDegree[node] == 0 {
			queue = append(queue, node)
		}
	}
	sort.Strings(queue) // Stable order

	var result []string
	processed := make(map[string]bool)

	for len(queue) > 0 {
		// Take from front
		node := queue[0]
		queue = queue[1:]

		if processed[node] {
			continue
		}
		processed[node] = true
		result = append(result, node)

		// Find nodes that depend on this node and decrement their in-degree
		for _, n := range nodes {
			if processed[n] {
				continue
			}
			for _, dep := range getDeps(n) {
				if dep == node {
					inDegree[n]--
					if inDegree[n] == 0 {
						queue = append(queue, n)
					}
					break
				}
			}
		}
		sort.Strings(queue) // Maintain stable order
	}

	// Handle cycles by adding remaining nodes
	for _, node := range nodes {
		if !processed[node] {
			result = append(result, node)
		}
	}

	return result
}
