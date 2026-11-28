package metrics

import (
	"go/ast"
	"go/parser"
	"go/token"
)

// GoComplexity returns the cyclomatic complexity of each function in a Go
// file: 1, plus 1 for each if, for, range, case (not default), comm clause, && and ||.
func GoComplexity(src []byte) (map[string]int, error) {
	// Purpose: parse a Go file and count branch points in each function.
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "source.go", src, 0)
	if err != nil {
		return nil, err
	}
	out := map[string]int{}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		out[funcName(fn)] = complexity(fn.Body)
	}
	return out, nil
}

// funcName names a method Type.Method and a function by its identifier.
func funcName(fn *ast.FuncDecl) string {
	// Purpose: build the map key for one function declaration.
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return fn.Name.Name
	}
	t := fn.Recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	name := ""
	switch v := t.(type) {
	case *ast.Ident:
		name = v.Name
	case *ast.IndexExpr:
		if id, ok := v.X.(*ast.Ident); ok {
			name = id.Name
		}
	case *ast.SelectorExpr:
		name = v.Sel.Name
	default:
		name = "?"
	}
	return name + "." + fn.Name.Name
}

// complexity counts one plus the branch points in one function body.
func complexity(body *ast.BlockStmt) int {
	// Purpose: walk a body without descending into nested function literals.
	n := 1
	ast.Inspect(body, func(node ast.Node) bool {
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}
		switch v := node.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.CommClause:
			n++
		case *ast.CaseClause:
			if len(v.List) > 0 {
				n++
			}
		case *ast.BinaryExpr:
			if v.Op == token.LAND || v.Op == token.LOR {
				n++
			}
		}
		return true
	})
	return n
}

// ComplexityAdded sums new minus old for every function in either map.
func ComplexityAdded(oldFuncs, newFuncs map[string]int) int {
	// Purpose: sum the complexity delta over the union of function names.
	names := map[string]bool{}
	for name := range oldFuncs {
		names[name] = true
	}
	for name := range newFuncs {
		names[name] = true
	}
	total := 0
	for name := range names {
		total += newFuncs[name] - oldFuncs[name]
	}
	return total
}
