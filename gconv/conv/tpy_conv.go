package conv

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

type typCtx struct {
	Ignore ignoreMap
	LIdent string
	RIdent string
}

func NewTpyCtx(ignoreMap ignoreMap, lname, rname string) *typCtx {
	return &typCtx{Ignore: ignoreMap, LIdent: lname, RIdent: rname}
}

func (tcc *typCtx) merge(tpyCtx *typCtx) *typCtx {
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

func (tcc *typCtx) isExist(name string) bool {
	return tcc.Ignore.exist(name)
}

func GenTpyConv(ctx *typCtx, lt types.Type, rt types.Type) (stmt []ast.Stmt) {
	if ctx == nil {
		ctx = &typCtx{}
	}

	switch x := lt.(type) {
	case *types.Basic:
		y, ok := rt.(*types.Basic)
		if !ok {
			return nil
		}
		return append(stmt, assignStmt(x, y, ctx.LIdent, ctx.RIdent))

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
			var xf, yf *types.Var

			xf = x.Field(i)
			if yf, ok = yMap[xf.Name()]; !ok {
				// NOTE[Dokiy] 2022/9/28: 记录未处理到字段，统一打印提示
				continue
			}

			var newCtx = (&typCtx{LIdent: xf.Name(), RIdent: yf.Name()}).merge(ctx)
			// Ignore filed
			if newCtx.isExist(newCtx.LIdent) {
				continue
			}

			// init pointer field
			switch xf.Type().(type) {
			case *types.Pointer:
				stmts = append(stmts, initVar(xf, newCtx.LIdent)...)
			}
			stmts = append(stmts, GenTpyConv(newCtx, xf.Type(), yf.Type())...)
		}
		return stmts

	case *types.Named:
		return GenTpyConv(ctx, x.Underlying(), underTpy(rt))

	case *types.Pointer:
		return GenTpyConv(ctx, x.Elem(), underTpy(rt))

	case *types.Slice, *types.Array:
		// TODO[Dokiy] 2022/9/28: to be continued!
		return nil
	//for i = 0; i < len(params); i++ {
	//	result[i].id = params[i].id
	//}
	//
	case *types.Map:
		// NOTE[Dokiy] 2022/9/29:
		return nil

	default:
	}
	// NOTE[Dokiy] 2022/9/29: 记录未处理到字段，统一打印提示
	panic("Unsupported type")
}

func assignStmt(lt *types.Basic, rt *types.Basic, lname string, rname string) ast.Stmt {
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
