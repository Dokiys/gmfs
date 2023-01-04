package conv

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

func assignStruct(typ types.Type, pkgAlias map[string]string, name string, kvExprs ...ast.Expr) []ast.Expr {
	var expr = &ast.UnaryExpr{}
	for {
		switch xx := typ.(type) {
		case *types.Pointer:
			typ = xx.Elem()
			expr.Op = token.AND
			continue
		case *types.Named:
			expr.Op = token.AND
			elts := make([]ast.Expr, 0, len(kvExprs))
			for _, e := range kvExprs {
				elts = append(elts, e)
			}
			if alias, ok := pkgAlias[xx.Obj().Pkg().Path()]; !ok {
				expr.X = &ast.CompositeLit{
					Type: ast.NewIdent(xx.Obj().Name()),
					Elts: elts,
				}
			} else {
				expr.X = &ast.CompositeLit{
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
			return []ast.Expr{&ast.KeyValueExpr{
				Key:   ast.NewIdent(name),
				Value: expr,
			}}
		default:
			return nil
		}
	}
}

func assignKV(key string, value string) ast.Expr {
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
		value = fmt.Sprintf("(%s)%s", keyTyp.String(), value)
		return &ast.CallExpr{
			Fun: &ast.Ident{
				Name: keyTyp.String(),
			},
			Args: []ast.Expr{ast.NewIdent(value)},
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
