package conv

import (
	"go/ast"
	"go/types"
)

type fnConv struct {
	panicStmt []ast.Stmt
}

// genFnConv just handle function which has one params, one results and panic a function in the end.
func genFnConv(info *types.Info, fn *ast.FuncDecl) (*fnConv, bool) {
	if len(fn.Type.Params.List) != 1 || len(fn.Type.Results.List) != 1 {
		return nil, false
	}

	var stmts []ast.Stmt
	for i, stmt := range fn.Body.List {
		switch stmt := stmt.(type) {
		case *ast.EmptyStmt:
			// Do nothing.
		case *ast.ReturnStmt:
			// Do nothing.
		case *ast.ExprStmt:
			call, ok := stmt.X.(*ast.CallExpr)
			if !ok {
				continue
			}

			// Handle last panic only.
			convObj := qualifiedIdentObject(info, call.Fun)
			if convObj != types.Universe.Lookup("panic") || i+1 != len(fn.Body.List) {
				continue
			}

			for _, arg := range call.Args {
				fl, ok := arg.(*ast.FuncLit)
				if !ok {
					continue
				}
				stmts = append(stmts, fl.Body.List...)
			}
			return &fnConv{panicStmt: stmts}, true
		}
	}

	return nil, false
}

func (c *fnConv) genConvStmt() []ast.Stmt {
	return c.panicStmt
}

func (c *fnConv) convFields(x *ast.FuncDecl) {
	// TODO[Dokiy] 2022/8/11: Conv Fields
}
