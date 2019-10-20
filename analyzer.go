package seams

import (
	"fmt"
	"go/ast"
	"go/types"
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

	inspect.Preorder([]ast.Node{
		(*ast.CallExpr)(nil),
	}, func(n ast.Node) {
		c := n.(*ast.CallExpr)
		file := pass.Fset.File(c.Pos())

		// Ignore tests.
		if strings.HasSuffix(file.Name(), "_test.go") {
			return
		}

		// Ignore go-build.
		if !strings.HasSuffix(file.Name(), ".go") {
			return
		}

		f := function(n, pass)
		if f == nil {
			return
		}

		pass.Report(analysis.Diagnostic{
			Pos:     c.Pos(),
			Message: fmt.Sprintf("call of untestable function/method: %s", f.FullName()),
		})

	})

	return nil, nil
}

func function(n ast.Node, pass *analysis.Pass) *types.Func {
	i := identifier(n)
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

func identifier(n ast.Node) *ast.Ident {
	c := n.(*ast.CallExpr)

	switch f := c.Fun.(type) {
	case *ast.Ident:
		return f
	case *ast.SelectorExpr:
		return f.Sel
	default:
		return nil
	}
}
