package conv

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

type varConv struct {
	lv    *types.Var
	rv    *types.Var
	lname string
	rname string
}

func newVarConv(lv *types.Var, rv *types.Var, lname string, rname string) *varConv {
	return &varConv{lv: lv, rv: rv, lname: lname, rname: rname}
}

func genVarConv(lv *types.Var, rv *types.Var, lname string, rname string) (stmt []ast.Stmt) {
	// TODO[Dokiy] 2022/8/12: 比较field, 生成 ast.stmt
	// NOTE[Dokiy] 2022/8/12: 处理dao.Item
	// NOTE[Dokiy] 2022/8/12: 处理[]*dao.Item
	// NOTE[Dokiy] 2022/8/12: 处理[]*dao.Item 与 []dao.Item

	if as := genAssignStmt(lv, rv, lname, rname); as != nil {
		return append(stmt, as)
	}

	switch x := underPointerTpy(lv.Type()).(type) {
	// int int 【x】
	// *dao.Item dao.Item 【x】
	//case *types.Basic, *types.Pointer:
	//	return append(stmt, genAssignStmt(lv, rv, lname, rname))
	case *types.Struct:
		for i := 0; i < x.NumFields(); i++ {
			//f := x.Field(i)
		}
	case *types.Slice, *types.Array:

	default:
		return nil
	}

	return nil
}

func genAssignStmt(lv *types.Var, rv *types.Var, lname string, rname string) ast.Stmt {
	lt, rt := underPointerTpy(lv.Type()), underPointerTpy(rv.Type())
	// Assign the same type.
	if lt == rt {
		return &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.SelectorExpr{
				X:   ast.NewIdent(lname),
				Sel: ast.NewIdent(lv.Name()),
			}},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{&ast.SelectorExpr{
				X:   ast.NewIdent(rname),
				Sel: ast.NewIdent(rv.Name()),
			}},
		}
	}

	// Assign different integer type, but can be converted
	if i, ok := lt.(*types.Basic); ok && i.Info() == types.IsInteger {
		rname = fmt.Sprintf("(%s)%s", i.Name(), rname)
		return &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.SelectorExpr{
				X:   ast.NewIdent(lname),
				Sel: ast.NewIdent(lv.Name()),
			}},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{&ast.SelectorExpr{
				X:   ast.NewIdent(rname),
				Sel: ast.NewIdent(rv.Name()),
			}},
		}
	}

	return nil
}

func pointerInitStmt(impAlias map[string]string, tpy *types.Pointer, name string) ast.Stmt {
	//{name} := new(tpy)
	// TODO[Dokiy] 2022/8/17: tpy.pkg 获取pkg name, 如果没获取到则取最后的名称
	// TODO[Dokiy] 2022/8/17: 当tpy为pointer时， new(typ.typeName)
	// 否则不做处理
	n, ok := tpy.Elem().(*types.Named)
	if !ok {
		panic("pointerInitStmt not *types.Named type")
	}
	_, ok = impAlias[n.Obj().Pkg().Path()]
	if !ok {
		// TODO[Dokiy] 2022/8/17: 获取到import最后到name
		return nil
	}

	// 同时获取type Name
	return &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(name)},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{ast.NewIdent(fmt.Sprintf("new(%s)", tpy.Underlying().String()))},
	}
}
