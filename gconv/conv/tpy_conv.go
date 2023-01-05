package conv

import (
	"go/ast"
	"go/types"
)

type typCtx struct {
	PkgAlias  map[string]string
	Ignore    ignoreMap
	KeyName   string
	ValueName string
}

func NewTpyCtx(pkgAlias map[string]string, ignoreMap ignoreMap, lname, rname string) *typCtx {
	return &typCtx{PkgAlias: pkgAlias, Ignore: ignoreMap, KeyName: lname, ValueName: rname}
}

func (tcc *typCtx) merge(tpyCtx *typCtx) *typCtx {
	if tpyCtx == nil {
		return tcc
	}

	tcc.Ignore = tpyCtx.Ignore
	tcc.PkgAlias = tpyCtx.PkgAlias
	tcc.KeyName = tcc.KeyName
	if len(tpyCtx.ValueName) > 0 {
		tcc.ValueName = tpyCtx.ValueName + "." + tcc.ValueName
	}

	return tcc
}

func (tcc *typCtx) shouldIgnore(name string) bool {
	return tcc.Ignore.exist(name)
}

func GenTpyConv(ctx *typCtx, lt types.Type, rt types.Type) (expr []ast.Expr) {
	if ctx == nil {
		ctx = &typCtx{}
	}

	switch x := lt.(type) {
	case *types.Basic:
		return append(expr, tryAssign(lt, rt, ctx.KeyName, ctx.ValueName))

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

		exprs := make([]ast.Expr, 0, x.NumFields())
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

			var newCtx = (&typCtx{KeyName: xVar.Name(), ValueName: yVar.Name()}).merge(ctx)
			if newCtx.shouldIgnore(newCtx.KeyName) {
				continue
			}

			// Assign same type field
			if types.IdenticalIgnoreTags(lt, rt) {
				exprs = append(exprs, kv(newCtx.KeyName, newCtx.ValueName))
				continue
			}

			exprs = append(exprs, GenTpyConv(newCtx, xVar.Type(), yVar.Type())...)
		}
		return exprs

	case *types.Named:
		// Assign same type field
		if types.IdenticalIgnoreTags(lt, rt) {
			return []ast.Expr{kv(ctx.KeyName, ctx.ValueName)}
		}

		return GenTpyConv(ctx, x.Underlying(), underTpy(rt))

	case *types.Pointer:
		// Assign same type field
		if types.IdenticalIgnoreTags(lt, rt) {
			return []ast.Expr{kv(ctx.KeyName, ctx.ValueName)}
		}

		// NOTE[Dokiy] 2023/1/4: **A unsupported
		return []ast.Expr{newStruct(x, ctx.PkgAlias, ctx.KeyName, GenTpyConv(ctx, x.Elem(), underTpy(rt))...)}

	case *types.Slice, *types.Array:
		// TODO[Dokiy] 2022/9/30:
		// for i = 0; i < len(params); i++ {
		//	result[i].id = params[i].id
		// }
		//
		return nil

	case *types.Map:
		if types.IdenticalIgnoreTags(lt, rt) {
			return []ast.Expr{kv(ctx.KeyName, ctx.ValueName)}
		}
		return nil

	default:
	}
	// NOTE[Dokiy] 2022/9/29: add_err
	panic("Unsupported type")
}
