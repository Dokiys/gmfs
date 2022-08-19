package conv

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

type varConv struct {
	ignore       ignoreMap
	ignorePrefix string
}

func newVarConv(ignoreMap ignoreMap, ignorePrefix string) *varConv {
	return &varConv{ignore: ignoreMap, ignorePrefix: ignorePrefix}
}

func (vc *varConv) genVarConv(lt types.Type, rt types.Type, lname string, rname string) (stmt []ast.Stmt) {
	// TODO[Dokiy] 2022/8/12: 比较field, 生成 ast.stmt
	//pMap, rFields := getFieldsMap(lt, emptyIgnoreMap), getFields(rt, vc.ignore)
	//for _, rf := range rFields {
	//	pf, ok := pMap[rf.Name()]
	//	if !ok || rf.Name() != pf.Name() {
	//		continue
	//	}
	//	// ignore = f.ignore[rf.Name()]
	//	stmt = append(stmt, vc.genVarConv(rf, pf, resultName, paramName)...)
	//}
	//
	//if as := genAssignStmt(lv, rv, lname, rname); as != nil {
	//	return append(stmt, as)
	//}

	if lt.String() == rt.String() {
		return append(stmt, &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(lname)},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent(rname)},
		})
	}

	switch lx := lt.(type) {
	// int int 【x】
	// *dao.Item dao.Item 【x】
	case *types.Basic:
		//if ok := tryType(rt); !ok {
		//	panic("err type")
		//}
		return append(stmt, genAssignStmt(lt, rt, lname, rname))

	case *types.Struct:
		_, ok := underPointerTpy(rt).(*types.Struct)
		if !ok {
			return nil
		}

		for i := 0; i < x.NumFields(); i++ {
			f := x.Field(i)
			panic(f)
		}
	case *types.Slice, *types.Array:
	//for i = 0; i < len(params); i++ {
	//	result[i].id = params[i].id
	//}
	//
	case *types.Named:
		//vc.genVarConv(lt, rt, lname, rname)

	default:
		return nil
	}

	return nil
}

func genAssignStmt(lt types.Type, rt types.Type, lname string, rname string) ast.Stmt {
	// Assign the same type.
	if lt.String() == rt.String() {
		return &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(lname)},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent(rname)},
		}
	}

	// Assign different integer type, but can be converted
	if i, ok := lt.(*types.Basic); ok && i.Info() == types.IsInteger {
		rname = fmt.Sprintf("(%s)%s", i.Name(), rname)
		return &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(lname)},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent(rname)},
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
