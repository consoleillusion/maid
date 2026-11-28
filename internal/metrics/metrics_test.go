package metrics

import (
	"reflect"
	"testing"
)

// scenario: dependencies added
func TestGoDeps(t *testing.T) {
	// Purpose: direct requirements are runtime, indirect ones are dev.
	oldMod := []byte("module x\n\ngo 1.24\n\nrequire example.com/a v1.0.0\n")
	newMod := []byte("module x\n\ngo 1.24\n\nrequire (\n\texample.com/a v1.0.0\n\texample.com/b v1.2.0\n\texample.com/c v0.1.0 // indirect\n)\n")
	runtime, dev, added, err := GoDeps(oldMod, newMod)
	if err != nil {
		t.Fatal(err)
	}
	if runtime != 1 || dev != 1 || !reflect.DeepEqual(added, []string{"example.com/b", "example.com/c"}) {
		t.Errorf("GoDeps = %d, %d, %v; want 1, 1, [example.com/b example.com/c]", runtime, dev, added)
	}
}

// scenario: require block and single-line require
func TestGoDepsRequireBlock(t *testing.T) {
	// Purpose: both require forms are counted.
	newMod := []byte("module x\n\ngo 1.24\n\nrequire example.com/one v1.0.0\n\nrequire (\n\texample.com/two v1.0.0\n\texample.com/three v1.0.0\n)\n")
	runtime, _, _, err := GoDeps(nil, newMod)
	if err != nil {
		t.Fatal(err)
	}
	if runtime != 3 {
		t.Errorf("runtimeAdded = %d, want 3", runtime)
	}
}

// scenario: complexity of a function
func TestGoComplexity(t *testing.T) {
	// Purpose: one if, one for and one && give complexity 4.
	src := []byte("package p\n\nfunc F(x int) int {\n\tif x > 0 && x < 10 {\n\t\treturn x\n\t}\n\tfor i := 0; i < x; i++ {\n\t\tx--\n\t}\n\treturn x\n}\n")
	got, err := GoComplexity(src)
	if err != nil {
		t.Fatal(err)
	}
	if got["F"] != 4 {
		t.Errorf("complexity of F = %d, want 4 (all: %v)", got["F"], got)
	}
}

// scenario: complexity of methods
func TestGoComplexityMethod(t *testing.T) {
	// Purpose: methods are named Type.Method.
	src := []byte("package p\n\ntype T struct{}\n\nfunc (T) M() {}\n")
	got, err := GoComplexity(src)
	if err != nil {
		t.Fatal(err)
	}
	if got["T.M"] != 1 {
		t.Errorf("complexity of T.M = %d, want 1 (all: %v)", got["T.M"], got)
	}
}

// scenario: complexity added
func TestComplexityAdded(t *testing.T) {
	// Purpose: the result is new minus old over functions in either map.
	oldFuncs := map[string]int{"A": 2, "B": 3}
	newFuncs := map[string]int{"A": 4, "C": 1}
	if got := ComplexityAdded(oldFuncs, newFuncs); got != 0 {
		t.Errorf("ComplexityAdded = %d, want 0", got)
	}
	if got := ComplexityAdded(map[string]int{}, map[string]int{"A": 3}); got != 3 {
		t.Errorf("ComplexityAdded for a new function = %d, want 3", got)
	}
}
