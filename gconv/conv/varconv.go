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

/*
	Var init stmt
*/
type varInitStmtOpt func(vs *varIniter)

func setSliceLen(len string) varInitStmtOpt {
	return func(vs *varIniter) {
		vs.len = len
	}
}

type varIniter struct {
	isNested bool
	impAlias map[string]string
	typ      types.Type
	len      string
}

func newVarIniter(impAlias map[string]string, typ types.Type, opts ...varInitStmtOpt) *varIniter {
	vi := &varIniter{impAlias: impAlias, typ: typ}
	for _, opt := range opts {
		opt(vi)
	}
	return vi
}

func (vi *varIniter) initIdent() string {
	return vi.init(vi.typ)
}

// NOTE[Dokiy] 2022/8/19: add test
func (vi *varIniter) init(typ types.Type) string {
	switch x := typ.(type) {
	case *types.Named:
		var brackets string
		if !vi.isNested {
			brackets = "{}"
		}

		alias, ok := vi.impAlias[x.Obj().Pkg().Path()]
		if !ok {
			return fmt.Sprintf("%s%s", x.Obj().Name(), brackets)
		}
		return fmt.Sprintf("%s.%s%s", alias, x.Obj().Name(), brackets)

	case *types.Pointer:
		if vi.isNested {
			return fmt.Sprintf("*%s", vi.init(x.Elem()))
		}
		vi.isNested = true
		return fmt.Sprintf("new(%s)", vi.init(x.Elem()))

	case *types.Slice:
		if vi.isNested {
			return fmt.Sprintf("[]%s", vi.init(x.Elem()))
		}
		vi.isNested = true
		str := fmt.Sprintf("make([]%s, %s)", vi.init(x.Elem()), vi.len)
		return str

	default:
		panic("varInitStmt unsupported type")
	}

	return ""
}
