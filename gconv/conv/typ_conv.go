package conv

import (
	"fmt"
	"go/types"
)

const (
	assignToken_DEFINE   = " :="
	assignToken_COLON    = ":"
	structTrailing_ENTER = "\n"
	structTrailing_COMMA = ","
)

type TypConvContext struct {
	keyName        string
	valueName      string
	assignToken    string
	structTrailing string

	pValueOnly bool
}

func NewTypCtx(keyName, valueName string) *TypConvContext {
	return &TypConvContext{
		keyName:        keyName,
		valueName:      valueName,
		assignToken:    assignToken_DEFINE,
		structTrailing: structTrailing_ENTER,
		pValueOnly:     false,
	}
}

func mergeCtx(tpyCtx *TypConvContext, keyName, valueName string) *TypConvContext {
	if tpyCtx == nil {
		return nil
	}

	keyName = keyName
	if len(tpyCtx.valueName) > 0 {
		valueName = tpyCtx.valueName + "." + valueName
	}

	return &TypConvContext{
		keyName:        keyName,
		valueName:      valueName,
		assignToken:    tpyCtx.assignToken,
		structTrailing: tpyCtx.structTrailing,
		pValueOnly:     tpyCtx.pValueOnly,
	}
}

type TypConvGen struct {
	g *gener

	pkgAlias map[string]string
	ignore   ignoreMap
	kt       types.Type
	vt       types.Type
}

func (tcg *TypConvGen) fork(kt, vt types.Type) *TypConvGen {
	// return tcg.forkWithGener(kt, vt, tcg.g)
	return &TypConvGen{
		g:        newGener(""),
		pkgAlias: tcg.pkgAlias,
		ignore:   tcg.ignore,
		kt:       kt,
		vt:       vt,
	}
}

func (tcg *TypConvGen) forkWithGener(kt, vt types.Type, g *gener) *TypConvGen {
	return &TypConvGen{
		g:        g,
		pkgAlias: tcg.pkgAlias,
		ignore:   tcg.ignore,
		kt:       kt,
		vt:       vt,
	}
}

func (tcg *TypConvGen) shouldIgnore(name string) bool {
	return tcg.ignore.exist(name)
}

func (tcg *TypConvGen) Gen(ctx *TypConvContext) {
	switch x := tcg.kt.(type) {
	case *types.Basic:
		if types.AssignableTo(tcg.vt, tcg.kt) {
			tcg.kv(ctx, ctx.keyName, ctx.valueName)
			return
		}

		// Assign different type which can be converted
		if types.ConvertibleTo(tcg.vt, tcg.kt) {
			tcg.kv(ctx, ctx.keyName, fmt.Sprintf("%s(%s)", tcg.kt.String(), ctx.valueName))
			return
		}
		// NOTE[Dokiy] 2022/9/30: add_err
		panic("Unsupported AssignStmt!")
		return

	case *types.Struct:
		// Assign symbol use colon in struct
		ctx.assignToken = assignToken_COLON
		ctx.structTrailing = structTrailing_COMMA + structTrailing_ENTER
		defer func() {
			ctx.assignToken = assignToken_DEFINE
			ctx.structTrailing = structTrailing_ENTER
		}()

		y, ok := tcg.vt.(*types.Struct)
		if !ok {
			return
		}

		yMap := make(map[string]*types.Var)
		for i := 0; i < y.NumFields(); i++ {
			if !y.Field(i).Exported() {
				continue
			}
			yMap[y.Field(i).Name()] = y.Field(i)
		}

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

			newCtx := mergeCtx(ctx, xVar.Name(), yVar.Name())
			newCtx.pValueOnly = true
			if tcg.shouldIgnore(newCtx.keyName) {
				continue
			}

			// Assign same type field
			if types.IdenticalIgnoreTags(tcg.kt, tcg.vt) {
				tcg.g.p("%s: %s,\n", newCtx.keyName, newCtx.valueName)
				// tcg.kv(newCtx, newCtx.keyName, newCtx.valueName)
				continue
			}

			// struct keep kv
			forkedTcg := tcg.forkWithGener(xVar.Type(), yVar.Type(), newGener(""))
			forkedTcg.Gen(newCtx)
			tcg.g.p("%s: %s,\n", newCtx.keyName, forkedTcg.g.string())
			// tcg.kv(ctx, newCtx.keyName, forkedTcg.g.string())
		}
		return

	case *types.Named:
		// Assign same type field
		if types.IdenticalIgnoreTags(tcg.kt, tcg.vt) {
			tcg.kv(ctx, ctx.keyName, ctx.valueName)
			return
		}

		// TODO[Dokiy] 2023/1/6: if vt is pointer, must consider nil
		var name = x.Obj().Name()
		if alias, ok := tcg.pkgAlias[x.Obj().Pkg().Path()]; ok {
			name = alias + "." + name
		}

		forkedTcg := tcg.forkWithGener(x.Underlying(), underTpy(tcg.vt), newGener(""))
		forkedTcg.Gen(ctx)
		tcg.kv(ctx, ctx.keyName, fmt.Sprintf("%s{\n%s}", name, forkedTcg.g.string()))
		// tcg.structDefine(x, forked
		// Tcg.g.string())

		return

	case *types.Pointer:
		// Assign same type field
		if types.IdenticalIgnoreTags(tcg.kt, tcg.vt) {
			tcg.kv(ctx, ctx.keyName, ctx.valueName)
			return
		}

		// TODO[Dokiy] 2023/1/6: if vt is pointer, must consider nil
		// NOTE[Dokiy] 2023/1/4: **A unsupported
		forkedTcg := tcg.fork(x.Elem(), underTpy(tcg.vt))
		forkedTcg.Gen(ctx)
		tcg.kv(ctx, ctx.keyName, "&"+forkedTcg.g.string())
		return

	case *types.Slice, *types.Array:
		// TODO[Dokiy] 2022/9/30:
		// for i = 0; i < len(params); i++ {
		//	result[i].id = params[i].id
		// }
		//
		return

	case *types.Map:
		if types.IdenticalIgnoreTags(tcg.kt, tcg.vt) {
			tcg.kv(ctx, ctx.keyName, ctx.valueName)
			return
		}
		return

	default:
	}
	// NOTE[Dokiy] 2022/9/29: add_err
	panic("Unsupported type")
}

func (tcg *TypConvGen) kv(ctx *TypConvContext, key string, value string) {
	if ctx.pValueOnly {
		tcg.g.p("%s", value)
	} else {
		tcg.g.p("%s%s %s%s", key, ctx.assignToken, value, ctx.structTrailing)
	}
	return
}
