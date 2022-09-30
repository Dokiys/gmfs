package conv

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

func initVar(obj types.Object, pkgAlias map[string]string, name string) []ast.Stmt {
	typ := obj.Type()
	ident := ""
	for {
		switch xx := typ.(type) {
		case *types.Pointer:
			typ = xx.Elem()
			ident = "&" + ident
			continue
		case *types.Named:
			if alias, ok := pkgAlias[xx.Obj().Pkg().Path()]; !ok {
				ident += fmt.Sprintf("%s%s", xx.Obj().Name(), "{}")
			} else {
				ident += fmt.Sprintf("%s.%s%s", alias, xx.Obj().Name(), "{}")
			}
			return []ast.Stmt{&ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent(name)},
				Tok: token.ASSIGN,
				// NOTE[Dokiy] 2022/9/30:
				Rhs: []ast.Expr{ast.NewIdent(ident)},
			}}
		default:
			return nil
		}
	}
}

func noneAssign(lname string, rname string) ast.Stmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(lname)},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{ast.NewIdent(rname)},
	}
}

func assign(lt *types.Basic, rt *types.Basic, lname string, rname string) ast.Stmt {
	// Assign the same type.
	if lt.String() == rt.String() {
		return &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(lname)},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent(rname)},
		}
	}

	// Assign different integer type, but can be converted
	if lt.Info()&types.IsInteger != 0 {
		rname = fmt.Sprintf("(%s)%s", lt.Name(), rname)
		return &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(lname)},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent(rname)},
		}
	}

	// NOTE[Dokiy] 2022/9/30: add_err
	panic("Unsupported AssignStmt!")
}
