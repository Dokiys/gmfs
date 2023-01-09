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
	kt        types.Type
	keyName   string
	vt        types.Type
	valueName string

	assignToken    string
	structTrailing string

	pValueOnly bool
}

func NewTypCtx(keyName, valueName string, kt, vt types.Type) *TypConvContext {
	return &TypConvContext{
		kt:             kt,
		keyName:        keyName,
		vt:             vt,
		valueName:      valueName,
		assignToken:    assignToken_DEFINE,
		structTrailing: structTrailing_ENTER,
		pValueOnly:     false,
	}
}

func (ctx *TypConvContext) mergeName(keyName, valueName string) *TypConvContext {
	if ctx == nil {
		return nil
	}

	ctx.keyName = keyName
	if len(ctx.valueName) > 0 {
		ctx.valueName = ctx.valueName + "." + valueName
	}

	return ctx
}

type TypConvGen struct {
	g *gener

	pkgAlias pkgAliasMap
	ignore   ignoreMap
}

func (tcg *TypConvGen) fork() *TypConvGen {
	return tcg.forkWithGener(newGener(""))
}

func (tcg *TypConvGen) forkWithGener(g *gener) *TypConvGen {
	return &TypConvGen{
		g:        g,
		pkgAlias: tcg.pkgAlias,
		ignore:   tcg.ignore,
	}
}

func (tcg *TypConvGen) shouldIgnore(name string) bool {
	return tcg.ignore.exist(name)
}

func (tcg *TypConvGen) Gen(ctx *TypConvContext) {
	switch x := ctx.kt.(type) {
	case *types.Basic:
		if types.AssignableTo(ctx.vt, ctx.kt) {
			tcg.kv(ctx, ctx.keyName, ctx.valueName)
			return
		}

		// Assign different type which can be converted
		if types.ConvertibleTo(ctx.vt, ctx.kt) {
			tcg.kv(ctx, ctx.keyName, fmt.Sprintf("%s(%s)", ctx.kt.String(), ctx.valueName))
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

		y, ok := ctx.vt.(*types.Struct)
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

			newCtx := (&TypConvContext{
				kt:             xVar.Type(),
				keyName:        ctx.keyName,
				vt:             yVar.Type(),
				valueName:      ctx.valueName,
				assignToken:    ctx.assignToken,
				structTrailing: ctx.structTrailing,
				pValueOnly:     true,
			}).mergeName(xVar.Name(), yVar.Name())
			if tcg.shouldIgnore(newCtx.keyName) {
				continue
			}

			// Assign same type field
			if types.IdenticalIgnoreTags(ctx.kt, ctx.vt) {
				tcg.g.p("%s: %s,\n", newCtx.keyName, newCtx.valueName)
				// tcg.kv(newCtx, newCtx.keyName, newCtx.valueName)
				continue
			}

			// struct keep kv
			forkedTcg := tcg.forkWithGener(newGener(""))
			forkedTcg.Gen(newCtx)
			tcg.g.p("%s: %s,\n", newCtx.keyName, forkedTcg.g.string())
			// tcg.kv(ctx, newCtx.keyName, forkedTcg.g.string())
		}
		return

	case *types.Named:
		// TODO[Dokiy] 2023/1/6: if vt is pointer, must consider nil
		var name = x.Obj().Name()
		if alias, ok := tcg.pkgAlias[x.Obj().Pkg().Path()]; ok {
			name = alias + "." + name
		}

		forkedTcg := tcg.forkWithGener(newGener(""))
		forkedTcg.Gen(&TypConvContext{
			kt:             x.Underlying(),
			keyName:        ctx.keyName,
			vt:             underTpy(ctx.vt),
			valueName:      ctx.valueName,
			assignToken:    ctx.assignToken,
			structTrailing: ctx.structTrailing,
			pValueOnly:     true,
		})
		tcg.kv(ctx, ctx.keyName, fmt.Sprintf("%s{\n%s}", name, forkedTcg.g.string()))

		return

	case *types.Pointer:
		// Assign same type field
		if types.IdenticalIgnoreTags(ctx.kt, ctx.vt) {
			tcg.kv(ctx, ctx.keyName, ctx.valueName)
			return
		}

		// TODO[Dokiy] 2023/1/9: to be continued!
		// TODO[Dokiy] 2023/1/6: if vt is pointer, must consider nil
		// NOTE[Dokiy] 2023/1/4: **A unsupported
		forkedTcg := tcg.fork()
		forkedTcg.Gen(&TypConvContext{
			kt:             x.Elem(),
			keyName:        ctx.keyName,
			vt:             ctx.vt,
			valueName:      ctx.valueName,
			assignToken:    ctx.assignToken,
			structTrailing: ctx.structTrailing,
			pValueOnly:     true,
		})
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
		if types.IdenticalIgnoreTags(ctx.kt, ctx.vt) {
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
