// Package metrics turns project files into score inputs.
// See design/method/scoring.md.
package metrics

// GoDeps compares two go.mod files. Direct requirements count as runtime,
// requirements marked // indirect and tool directives count as dev.
func GoDeps(oldMod, newMod []byte) (runtimeAdded, devAdded int, added []string, err error) {
	// Purpose: stub; the code work item implements it.
	return 0, 0, nil, nil
}
