package conv

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

func initVar(typ types.Type, pkgAlias map[string]string, name string) ast.Stmt {
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
			return &ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent(name)},
				Tok: token.ASSIGN,
				// NOTE[Dokiy] 2022/9/30:
				Rhs: []ast.Expr{ast.NewIdent(ident)},
			}
		default:
			return nil
		}
	}
}

func assgin(assigned string, assign string) ast.Stmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(assigned)},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{ast.NewIdent(assign)},
	}
}

func tryAssign(assignedTyp types.Type, assignTyp types.Type, assigned string, assign string) ast.Stmt {
	if types.AssignableTo(assignTyp, assignedTyp) {
		return &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(assigned)},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent(assign)},
		}
	}

	// Assign different type which can be converted
	if types.ConvertibleTo(assignTyp, assignedTyp) {
		assign = fmt.Sprintf("(%s)%s", assignedTyp.String(), assign)
		return &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(assigned)},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent(assign)},
		}
	}

	// NOTE[Dokiy] 2022/9/30: add_err
	panic("Unsupported AssignStmt!")
}

func setField(assignedTyp types.Type, assignTyp types.Type, assigned string, assign string) ast.Stmt {
	if types.AssignableTo(assignTyp, assignedTyp) {
		return &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(assigned)},
			Tok: token.COLON,
			Rhs: []ast.Expr{ast.NewIdent(assign)},
		}
	}

	// Assign different type which can be converted
	if types.ConvertibleTo(assignTyp, assignedTyp) {
		assign = fmt.Sprintf("(%s)%s", assignedTyp.String(), assign)
		return &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(assigned)},
			Tok: token.COLON,
			Rhs: []ast.Expr{ast.NewIdent(assign)},
		}
	}

	// NOTE[Dokiy] 2022/9/30: add_err
	panic("Unsupported AssignStmt!")
}
