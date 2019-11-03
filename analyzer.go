package seams

import (
	"fmt"
	"go/ast"
	"go/types"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:             "seams",
	Doc:              `find function/method calls which cannot be replaced by test doubles.`,
	Requires:         []*analysis.Analyzer{inspect.Analyzer},
	Run:              run,
	RunDespiteErrors: true,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	inspect.Nodes([]ast.Node{
		(*ast.File)(nil),
		(*ast.CallExpr)(nil),
	}, func(n ast.Node, push bool) bool {
		if !push {
			return false
		}

		switch n := n.(type) {
		case *ast.File:
			f := pass.Fset.File(n.Pos()).Name()
			if strings.HasSuffix(f, "_test.go") || !strings.HasSuffix(f, ".go") {
				return false
			}
			return !generated(n)
		case *ast.CallExpr:
			f := function(n, pass)
			if f == nil {
				break
			}

			pass.Report(analysis.Diagnostic{
				Pos:     n.Pos(),
				Message: fmt.Sprintf("untestable function/method call: %s", f.FullName()),
			})
		}

		return true
	})

	return nil, nil
}

func function(c *ast.CallExpr, pass *analysis.Pass) *types.Func {
	i := identifier(c)
	if i == nil {
		return nil
	}

	o, ok := pass.TypesInfo.Uses[i]
	if !ok {
		return nil
	}

	// Ignore global functions or functions/methods within the same package.
	if pkg := o.Pkg(); pkg == nil || pkg.Path() == pass.Pkg.Path() {
		return nil
	}

	// Ignore non-function calls e.g. type conversions or function variable calls.
	f, ok := o.(*types.Func)
	if !ok {
		return nil
	}

	s := f.Type().(*types.Signature)

	r := s.Recv()
	if r == nil { // function
		return f
	}

	// method

	// Ignore method calls on interfaces because we can mock them.
	if _, ok := r.Type().(*types.Interface); ok {
		return nil
	}

	return f
}

func identifier(c *ast.CallExpr) *ast.Ident {
	switch f := c.Fun.(type) {
	case *ast.Ident:
		return f
	case *ast.SelectorExpr:
		return f.Sel
	default:
		return nil
	}
}

// https://github.com/golang/go/issues/13560#issuecomment-288457920
var generatedPattern = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)

func generated(f *ast.File) bool {
	for _, c := range f.Comments {
		for _, l := range c.List {
			if generatedPattern.MatchString(l.Text) {
				return true
			}
		}
	}
	return false
}
