package conv

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

// TODO[Dokiy] 2022/9/28: modified name
type tpyConvCtx struct {
	ignore ignoreMap
	lname  string
	rname  string
}

func newTpyConvCtx(ignoreMap ignoreMap, lname, rname string) *tpyConvCtx {
	return &tpyConvCtx{ignore: ignoreMap, lname: lname, rname: rname}
}

func (tcc *tpyConvCtx) getLname() string {
	return tcc.lname
}

func (tcc *tpyConvCtx) getRname() string {
	return tcc.rname
}

func (tcc *tpyConvCtx) setLname(name string) {
	tcc.lname = name
}

func (tcc *tpyConvCtx) setRname(name string) {
	tcc.rname = name
}

func (tcc *tpyConvCtx) isExist(name string) bool {
	return tcc.ignore.exist(name)
}

func (tcc *tpyConvCtx) clone() tpyConvCtx {
	return *newTpyConvCtx(tcc.ignore, tcc.getLname(), tcc.getRname())
}

func genTpyConv(ctx tpyConvCtx, lt types.Type, rt types.Type) (stmt []ast.Stmt) {
	if lt.String() == rt.String() {
		return append(stmt, &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(ctx.getLname())},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent(ctx.getRname())},
		})
	}

	switch x := lt.(type) {
	case *types.Basic:
		y, ok := rt.(*types.Basic)
		if !ok {
			return nil
		}
		return append(stmt, genAssignStmt(x, y, ctx.getLname(), ctx.getRname()))

	case *types.Struct:
		y, ok := rt.(*types.Struct)
		if !ok {
			return nil
		}

		yMap := make(map[string]*types.Var)
		for i := 0; i < y.NumFields(); i++ {
			yMap[y.Field(i).Name()] = y.Field(i)
		}

		stmts := make([]ast.Stmt, 0, x.NumFields())
		for i := 0; i < x.NumFields(); i++ {
			xf := x.Field(i)

			yf, ok := yMap[xf.Name()]
			if !ok {
				// NOTE[Dokiy] 2022/9/28: 记录未处理到字段，统一打印提示
				continue
			}

			newCtx := ctx.clone()
			newCtx.setLname(newCtx.getLname() + "." + xf.Name())
			newCtx.setRname(newCtx.getRname() + "." + yf.Name())
			// Ignore filed
			if newCtx.isExist(newCtx.getLname()) {
				continue
			}

			stmts = append(stmts, genTpyConv(newCtx, xf.Type(), yf.Type())...)
		}
		return stmts

	case *types.Slice, *types.Array:
		// TODO[Dokiy] 2022/9/28: to be continued!
		return nil
	//for i = 0; i < len(params); i++ {
	//	result[i].id = params[i].id
	//}
	//
	case *types.Named:
		if y, ok := rt.(*types.Named); ok {
			return genTpyConv(ctx, x.Underlying(), y.Underlying())
		}
		return genTpyConv(ctx, x.Underlying(), rt)

	case *types.Pointer:
		if y, ok := rt.(*types.Pointer); ok {
			return genTpyConv(ctx, x.Elem(), y.Elem())
		}
		return genTpyConv(ctx, x.Elem(), rt)

	default:
	}
	panic("Unsupported type")
}

func genAssignStmt(lt *types.Basic, rt *types.Basic, lname string, rname string) ast.Stmt {
	// Assign the same type.
	if lt.String() == rt.String() {
		return &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(lname)},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent(rname)},
		}
	}

	// Assign different integer type, but can be converted
	if lt.Info() == types.IsInteger {
		rname = fmt.Sprintf("(%s)%s", lt.Name(), rname)
		return &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(lname)},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent(rname)},
		}
	}

	panic("")
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
