package conv

import (
	"go/ast"
	"go/types"
)

type typCtx struct {
	PkgAlias map[string]string
	Ignore   ignoreMap
	LIdent   string
	RIdent   string
}

func NewTpyCtx(pkgAlias map[string]string, ignoreMap ignoreMap, lname, rname string) *typCtx {
	return &typCtx{PkgAlias: pkgAlias, Ignore: ignoreMap, LIdent: lname, RIdent: rname}
}

func (tcc *typCtx) merge(tpyCtx *typCtx) *typCtx {
	if tpyCtx == nil {
		return tcc
	}

	tcc.Ignore = tpyCtx.Ignore
	tcc.PkgAlias = tpyCtx.PkgAlias
	if len(tpyCtx.LIdent) > 0 {
		tcc.LIdent = tpyCtx.LIdent + "." + tcc.LIdent
	}
	if len(tpyCtx.RIdent) > 0 {
		tcc.RIdent = tpyCtx.RIdent + "." + tcc.RIdent
	}

	return tcc
}

func (tcc *typCtx) shouldIgnore(name string) bool {
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
		return append(stmt, assign(x, y, ctx.LIdent, ctx.RIdent))

	case *types.Struct:
		y, ok := rt.(*types.Struct)
		if !ok {
			return nil
		}

		yMap := make(map[string]*types.Var)
		for i := 0; i < y.NumFields(); i++ {
			if !y.Field(i).Exported() {
				continue
			}
			yMap[y.Field(i).Name()] = y.Field(i)
		}

		stmts := make([]ast.Stmt, 0, x.NumFields())
		for i := 0; i < x.NumFields(); i++ {
			xVar := x.Field(i)
			if !xVar.Exported() {
				continue
			}
			yVar, ok := yMap[xVar.Name()]
			if !ok {
				// NOTE[Dokiy] 2022/9/28: add_err
				continue
			}

			var newCtx = (&typCtx{LIdent: xVar.Name(), RIdent: yVar.Name()}).merge(ctx)
			if newCtx.shouldIgnore(newCtx.LIdent) {
				continue
			}

			// Handle completely same type field
			if xVar.String() == yVar.String() {
				stmts = append(stmts, noneAssign(newCtx.LIdent, newCtx.RIdent))
				continue
			}

			switch xVar.Type().(type) {
			case *types.Pointer:
				// NOTE[Dokiy] 2022/9/30:
				// Need keep x and y isolated
				//	if xVar.Name() == yVar.Name() && xVar.Type().String() == types.NewPointer(yVar.Type()).String() {
				//		stmts = append(stmts, noneAssign(ctx.LIdent, token.AND.String()+ctx.RIdent))
				//		continue
				//	}
				if stmt := initVar(xVar, newCtx.PkgAlias, newCtx.LIdent); stmt != nil {
					stmts = append(stmts, stmt)
				}

			case *types.Named:
				// NOTE[Dokiy] 2022/9/30:
				// Need keep x and y isolated
				// 	if xf.Name() == yf.Name() && types.NewPointer(x).String() == yf.Type().String() {
				// 		stmts = append(stmts, assignStmt(nilBasic, nilBasic, newCtx.LIdent, token.MUL.String()+newCtx.RIdent))
				// 		continue
				// 	}
			}

			stmts = append(stmts, GenTpyConv(newCtx, xVar.Type(), yVar.Type())...)
		}
		return stmts

	case *types.Named:
		return GenTpyConv(ctx, x.Underlying(), underTpy(rt))

	case *types.Pointer:
		return GenTpyConv(ctx, x.Elem(), underTpy(rt))

	case *types.Slice, *types.Array:
		// TODO[Dokiy] 2022/9/30:
		//for i = 0; i < len(params); i++ {
		//	result[i].id = params[i].id
		//}
		//
		return nil

	case *types.Map:
		// NOTE[Dokiy] 2022/9/29: v0.2
		if lt.String() == rt.String() {
			return []ast.Stmt{noneAssign(ctx.LIdent, ctx.RIdent)}
		}
		return nil

	default:
	}
	// NOTE[Dokiy] 2022/9/29: 记录未处理到字段，统一打印提示
	panic("Unsupported type")
}

func tryInit(newCtx *typCtx, xVar, yVar *types.Var) ast.Stmt {
	switch xVar.Type().(type) {
	case *types.Pointer:
		// NOTE[Dokiy] 2022/9/30:
		// Need keep x and y isolated
		//	if xVar.Name() == yVar.Name() && xVar.Type().String() == types.NewPointer(yVar.Type()).String() {
		//		stmts = append(stmts, noneAssign(ctx.LIdent, token.AND.String()+ctx.RIdent))
		//		continue
		//	}
		return initVar(xVar, newCtx.PkgAlias, newCtx.LIdent)

	case *types.Named:
		// NOTE[Dokiy] 2022/9/30:
		// Need keep x and y isolated
		// 	if xf.Name() == yf.Name() && types.NewPointer(x).String() == yf.Type().String() {
		// 		stmts = append(stmts, assignStmt(nilBasic, nilBasic, newCtx.LIdent, token.MUL.String()+newCtx.RIdent))
		// 		continue
		// 	}
	}

	return nil
}
