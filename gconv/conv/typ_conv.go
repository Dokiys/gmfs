package conv

import (
	"fmt"
	"go/types"
)

const (
	assignToken_DEFINE   = " :="
	assignToken_AND      = ":"
	structTrailing_Enter = "\n"
	structTrailing_Comma = ","
)

type TypConvContext struct {
	keyName        string
	valueName      string
	assignToken    string
	structTrailing string
}

func NewTypCtx(keyName, valueName string) *TypConvContext {
	return &TypConvContext{
		keyName:        keyName,
		valueName:      valueName,
		assignToken:    assignToken_DEFINE,
		structTrailing: structTrailing_Enter,
	}
}

func (tcc *TypConvContext) merge(tpyCtx *TypConvContext) *TypConvContext {
	if tpyCtx == nil {
		return tcc
	}

	tcc.keyName = tcc.keyName
	if len(tpyCtx.valueName) > 0 {
		tcc.valueName = tpyCtx.valueName + "." + tcc.valueName
	}

	return tcc
}

type TypConvGen struct {
	Ctx *TypConvContext
	g   *gener

	pkgAlias map[string]string
	ignore   ignoreMap
	kt       types.Type
	vt       types.Type
}

func (tcg *TypConvGen) fork(ctx *TypConvContext, kt, vt types.Type) *TypConvGen {
	ctx.assignToken = assignToken_AND
	ctx.structTrailing = structTrailing_Comma
	return tcg.forkWithGener(ctx, kt, vt, tcg.g)
}

func (tcg *TypConvGen) forkWithGener(ctx *TypConvContext, kt, vt types.Type, g *gener) *TypConvGen {
	return &TypConvGen{
		Ctx:      ctx,
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

func (tcg *TypConvGen) gen() {
	switch x := tcg.kt.(type) {
	case *types.Basic:
		if types.AssignableTo(tcg.vt, tcg.kt) {
			tcg.kv(tcg.Ctx.keyName, tcg.Ctx.valueName)
			return
		}

		// Assign different type which can be converted
		if types.ConvertibleTo(tcg.vt, tcg.kt) {
			tcg.kv(tcg.Ctx.keyName, fmt.Sprintf("%s(%s)", tcg.kt.String(), tcg.Ctx.valueName))
			return
		}
		// NOTE[Dokiy] 2022/9/30: add_err
		panic("Unsupported AssignStmt!")
		return

	case *types.Struct:
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

			var newCtx = (&TypConvContext{keyName: xVar.Name(), valueName: yVar.Name()}).merge(tcg.Ctx)
			if tcg.shouldIgnore(newCtx.keyName) {
				continue
			}

			// Assign same type field
			if types.IdenticalIgnoreTags(tcg.kt, tcg.vt) {
				tcg.kv(newCtx.keyName, newCtx.valueName)
				continue
			}

			tcg.fork(newCtx, xVar.Type(), yVar.Type()).gen()
		}
		return

	case *types.Named:
		// Assign same type field
		if types.IdenticalIgnoreTags(tcg.kt, tcg.vt) {
			tcg.kv(tcg.Ctx.keyName, tcg.Ctx.valueName)
			return
		}

		pointerTcg := tcg.forkWithGener(tcg.Ctx, x.Underlying(), underTpy(tcg.vt), newGener(prefix_4Tab))
		pointerTcg.gen()
		tcg.structDefine(x, tcg.Ctx.keyName, pointerTcg.g.string())

		return

	case *types.Pointer:
		// Assign same type field
		if types.IdenticalIgnoreTags(tcg.kt, tcg.vt) {
			tcg.kv(tcg.Ctx.keyName, tcg.Ctx.valueName)
			return
		}

		// NOTE[Dokiy] 2023/1/4: **A unsupported
		pointerTcg := tcg.forkWithGener(tcg.Ctx, x.Elem(), underTpy(tcg.vt), newGener(""))
		pointerTcg.gen()
		tcg.structDefine(x, tcg.Ctx.keyName, pointerTcg.g.string())
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
			tcg.kv(tcg.Ctx.keyName, tcg.Ctx.valueName)
			return
		}
		return

	default:
	}
	// NOTE[Dokiy] 2022/9/29: add_err
	panic("Unsupported type")
}

func (tcg *TypConvGen) structDefine(typ types.Type, keyName, content string) {
	if content != "" {
		content = "\n" + content
	}

	tcg.g.p("%s%s ", keyName, tcg.Ctx.assignToken)
	for {
		switch xx := typ.(type) {
		case *types.Pointer:
			tcg.g.p("&")
			typ = xx.Elem()
			continue

		case *types.Named:
			if alias, ok := tcg.pkgAlias[xx.Obj().Pkg().Path()]; ok {
				tcg.g.p("%s.%s{%s}\n", alias, xx.Obj().Name(), content)
				return
			} else {
				tcg.g.p("%s{%s}\n", xx.Obj().Name(), content)
				return
			}
		default:
			return
		}
	}

}

func (tcg *TypConvGen) kv(key string, value string) {
	tcg.g.p("%s%s %s,\n", key, tcg.Ctx.assignToken, value)
	return
}
