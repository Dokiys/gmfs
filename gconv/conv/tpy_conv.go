package conv

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

type TypConvCtx struct {
	Ignore ignoreMap
	LIdent string
	RIdent string
}

func NewTpyConvCtx(ignoreMap ignoreMap, lname, rname string) *TypConvCtx {
	return &TypConvCtx{Ignore: ignoreMap, LIdent: lname, RIdent: rname}
}

func (tcc *TypConvCtx) Merge(tpyCtx *TypConvCtx) *TypConvCtx {
	if tpyCtx == nil {
		return tcc
	}

	tcc.Ignore = tpyCtx.Ignore
	if len(tpyCtx.LIdent) > 0 {
		tcc.LIdent = tpyCtx.LIdent + "." + tcc.LIdent
	}
	if len(tpyCtx.RIdent) > 0 {
		tcc.RIdent = tpyCtx.RIdent + "." + tcc.RIdent
	}

	return tcc
}

func (tcc *TypConvCtx) IsExist(name string) bool {
	return tcc.Ignore.exist(name)
}

func GenTpyConv(tpyCtx *TypConvCtx, lt types.Type, rt types.Type) (stmt []ast.Stmt) {
	if lt.String() == rt.String() {
		return append(stmt, &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(tpyCtx.LIdent)},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent(tpyCtx.RIdent)},
		})
	}

	switch x := lt.(type) {
	case *types.Basic:
		y, ok := rt.(*types.Basic)
		if !ok {
			return nil
		}
		return append(stmt, genAssignStmt(x, y, tpyCtx.LIdent, tpyCtx.RIdent))

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

			var newCtx = (&TypConvCtx{LIdent: xf.Name(), RIdent: yf.Name()}).Merge(tpyCtx)
			// Ignore filed
			if newCtx.IsExist(newCtx.LIdent) {
				continue
			}

			stmts = append(stmts, GenTpyConv(newCtx, xf.Type(), yf.Type())...)
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
		//if y, ok := rt.(*types.Pointer); ok {
		//	return GenTpyConv(tpyCtx, x.Underlying(), y.Elem())
		//}
		//if y, ok := rt.(*types.Named); ok {
		//	return GenTpyConv(tpyCtx, x.Underlying(), y.Underlying())
		//}
		return GenTpyConv(tpyCtx, x.Underlying(), underTpy(rt))

	case *types.Pointer:
		//if y, ok := rt.(*types.Pointer); ok {
		//	return GenTpyConv(tpyCtx, x.Elem(), y.Elem())
		//}
		//if y, ok := rt.(*types.Named); ok {
		//	return GenTpyConv(tpyCtx, x.Elem(), y.Underlying())
		//}
		return GenTpyConv(tpyCtx, x.Elem(), underTpy(rt))

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

	panic("Unsupported AssignStmt!")
}
