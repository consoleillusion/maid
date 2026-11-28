package metrics

// GoComplexity returns the cyclomatic complexity of each function in a Go
// file: 1, plus 1 for each if, for, range, case (not default), comm clause, && and ||.
func GoComplexity(src []byte) (map[string]int, error) {
	// Purpose: stub; the code work item implements it.
	return nil, nil
}

// ComplexityAdded sums new minus old for every function in either map.
func ComplexityAdded(oldFuncs, newFuncs map[string]int) int {
	// Purpose: stub; the code work item implements it.
	return 0
}
