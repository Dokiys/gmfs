package conv

import (
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

func genVarConv(lv *types.Var, rv *types.Var, lname string, rname string) ([]ast.Stmt, bool) {
	if rv.Name() != lv.Name() {
		return nil, false
	}

	// TODO[Dokiy] 2022/8/12: 比较field, 生成 ast.stmt
	// NOTE[Dokiy] 2022/8/12: 处理dao.Item
	// NOTE[Dokiy] 2022/8/12: 处理[]*dao.Item
	// NOTE[Dokiy] 2022/8/12: 处理[]*dao.Item 与 []dao.Item

	// int int 【x】
	// *dao.Item dao.Item 【x】
	if tpy, ok := genAssignStmt(lv, rv, lname, rname); ok {
		return tpy, true
	}

	switch lv.Type().(type) {
	//case *types.Pointer, *types.Basic:
	//	return genSameTpy()
	case *types.Struct:
		//for _, i := range x.NumFields() {
		//X:   ast.NewIdent("a." + lname),
		//
		//}
	case *types.Slice, *types.Array:

	}

	return nil, false
}

func genAssignStmt(lv *types.Var, rv *types.Var, lname string, rname string) (stmt []ast.Stmt, ok bool) {
	lt, rt := underPointerTpy(lv.Type()), underPointerTpy(rv.Type())
	if lt == rt {
		stmt = append(stmt, &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.SelectorExpr{
				X:   ast.NewIdent(lname),
				Sel: ast.NewIdent(lv.Name()),
			}},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{&ast.SelectorExpr{
				X:   ast.NewIdent(rname),
				Sel: ast.NewIdent(rv.Name()),
			}},
		})
		return stmt, true
	}

	// TODO[Dokiy] 2022/8/16: 1. int conv
	return nil, false
}
