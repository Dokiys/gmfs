package conv

import (
	"go/ast"
	"go/types"
)

type typCtx struct {
	PkgAlias      map[string]string
	Ignore        ignoreMap
	AssignedIdent string
	AssignIdent   string
}

func NewTpyCtx(pkgAlias map[string]string, ignoreMap ignoreMap, lname, rname string) *typCtx {
	return &typCtx{PkgAlias: pkgAlias, Ignore: ignoreMap, AssignedIdent: lname, AssignIdent: rname}
}

func (tcc *typCtx) merge(tpyCtx *typCtx) *typCtx {
	if tpyCtx == nil {
		return tcc
	}

	tcc.Ignore = tpyCtx.Ignore
	tcc.PkgAlias = tpyCtx.PkgAlias
	if len(tpyCtx.AssignedIdent) > 0 {
		tcc.AssignedIdent = tpyCtx.AssignedIdent + "." + tcc.AssignedIdent
	}
	if len(tpyCtx.AssignIdent) > 0 {
		tcc.AssignIdent = tpyCtx.AssignIdent + "." + tcc.AssignIdent
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
		return append(stmt, tryAssign(lt, rt, ctx.AssignedIdent, ctx.AssignIdent))

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

			var newCtx = (&typCtx{AssignedIdent: xVar.Name(), AssignIdent: yVar.Name()}).merge(ctx)
			if newCtx.shouldIgnore(newCtx.AssignedIdent) {
				continue
			}

			// Assign same type field
			if types.IdenticalIgnoreTags(lt, rt) {
				stmts = append(stmts, assgin(newCtx.AssignedIdent, newCtx.AssignIdent))
				continue
			}

			stmts = append(stmts, GenTpyConv(newCtx, xVar.Type(), yVar.Type())...)
		}
		return stmts

	case *types.Named:
		// Assign same type field
		if types.IdenticalIgnoreTags(lt, rt) {
			return []ast.Stmt{assgin(ctx.AssignedIdent, ctx.AssignIdent)}
		}

		return GenTpyConv(ctx, x.Underlying(), underTpy(rt))

	case *types.Pointer:
		stmts := make([]ast.Stmt, 0)
		// Assign same type field
		if types.IdenticalIgnoreTags(lt, rt) {
			return []ast.Stmt{assgin(ctx.AssignedIdent, ctx.AssignIdent)}
		}

		// NOTE[Dokiy] 2023/1/4: **A unsupported
		if stmt := initVar(x, ctx.PkgAlias, ctx.AssignedIdent); stmt != nil {
			stmts = append(stmts, stmt)
		}
		return append(stmts, GenTpyConv(ctx, x.Elem(), underTpy(rt))...)

	case *types.Slice, *types.Array:
		// TODO[Dokiy] 2022/9/30:
		// for i = 0; i < len(params); i++ {
		//	result[i].id = params[i].id
		// }
		//
		return nil

	case *types.Map:
		if types.IdenticalIgnoreTags(lt, rt) {
			return []ast.Stmt{assgin(ctx.AssignedIdent, ctx.AssignIdent)}
		}
		return nil

	default:
	}
	// NOTE[Dokiy] 2022/9/29: add_err
	panic("Unsupported type")
}
