package conv

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

func newStruct(typ types.Type, pkgAlias map[string]string, name string, kvExpr ...ast.Expr) ast.Expr {
	switch xx := typ.(type) {
	case *types.Pointer:
		return &ast.UnaryExpr{
			Op: token.AND,
			X:  newStruct(xx.Elem(), pkgAlias, name, kvExpr...),
		}
	case *types.Named:
		elts := make([]ast.Expr, 0, len(kvExpr))
		for _, e := range kvExpr {
			elts = append(elts, e)
		}
		if alias, ok := pkgAlias[xx.Obj().Pkg().Path()]; !ok {
			return &ast.CompositeLit{
				Type: ast.NewIdent(xx.Obj().Name()),
				Elts: elts,
			}
		} else {
			return &ast.CompositeLit{
				Type: &ast.SelectorExpr{
					X: &ast.Ident{
						Name: alias,
					},
					Sel: &ast.Ident{
						Name: xx.Obj().Name(),
					},
				},
				Elts: elts,
			}
		}
	default:
		return nil
	}
}

func kv(key string, value string) ast.Expr {
	return &ast.KeyValueExpr{
		Key:   ast.NewIdent(key),
		Value: ast.NewIdent(value),
	}
}

func tryAssign(keyTyp types.Type, valueType types.Type, key string, value string) ast.Expr {
	if types.AssignableTo(valueType, keyTyp) {
		return &ast.KeyValueExpr{
			Key:   ast.NewIdent(key),
			Value: ast.NewIdent(value),
		}
	}

	// Assign different type which can be converted
	if types.ConvertibleTo(valueType, keyTyp) {
		return &ast.KeyValueExpr{
			Key: ast.NewIdent(key),
			Value: &ast.CallExpr{
				Fun: &ast.Ident{
					Name: keyTyp.String(),
				},
				Args: []ast.Expr{ast.NewIdent(value)},
			},
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
