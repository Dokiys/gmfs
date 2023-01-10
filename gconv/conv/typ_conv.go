package conv

import (
	"fmt"
	"go/types"
	"strings"
)

const (
	assignToken_Assign   = " ="
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
	pg *gener
	g  *gener

	pkgAlias pkgAliasMap
	ignore   ignoreMap
}

func (tcg *TypConvGen) fork() *TypConvGen {
	return &TypConvGen{
		pg:       newGener(),
		g:        newGener(),
		pkgAlias: tcg.pkgAlias,
		ignore:   tcg.ignore,
	}
}

func (tcg *TypConvGen) shouldIgnore(name string) bool {
	return tcg.ignore.exist(name)
}

func (tcg *TypConvGen) Print() string {
	return tcg.pgPrint() + "\n" + tcg.gPrint()
}

func (tcg *TypConvGen) pgPrint() string {
	return tcg.pg.string()
}

func (tcg *TypConvGen) gPrint() string {
	return tcg.g.string()
}

func (tcg *TypConvGen) aliaName(typ types.Type) string {
	x, ok := typ.(*types.Named)
	if !ok {
		return strings.TrimLeft(typ.String(), ".")
	}

	var name = x.Obj().Name()
	if alias, ok := tcg.pkgAlias[x.Obj().Pkg().Path()]; ok {
		name = alias + "." + name
	}
	return name
}

func (tcg *TypConvGen) Gen(ctx *TypConvContext) {
	switch x := ctx.kt.(type) {
	case *types.Basic:
		if types.AssignableTo(ctx.vt, ctx.kt) {
			kv(tcg.g, ctx, ctx.keyName, ctx.valueName)
			return
		}

		// Assign different type which can be converted
		if types.ConvertibleTo(ctx.vt, ctx.kt) {
			kv(tcg.g, ctx, ctx.keyName, fmt.Sprintf("%s(%s)", ctx.kt.String(), ctx.valueName))
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
				continue
			}

			forkedTcg := tcg.fork()
			forkedTcg.Gen(newCtx)
			tcg.g.p("%s: %s,\n", newCtx.keyName, forkedTcg.g.string())
			tcg.pg.p(forkedTcg.pg.string())
		}
		return

	case *types.Named:
		// If vt is Pointer, check value is not nil.
		if _, ok := ctx.vt.(*types.Pointer); ok {
			// Prepare assign
			// TODO[Dokiy] 2023/1/9: to be continued!
			name := ctx.keyName + tcg.aliaName(x)
			kv(tcg.g, ctx, ctx.keyName, name)

			forkedTcg := tcg.fork()
			forkedTcg.Gen(&TypConvContext{
				kt:             x,
				keyName:        name,
				vt:             underTpy(ctx.vt),
				valueName:      ctx.valueName,
				assignToken:    assignToken_Assign,
				structTrailing: structTrailing_ENTER,
				pValueOnly:     false,
			})
			tcg.pg.p("var %s %s\n", name, tcg.aliaName(x))
			tcg.pg.p("if %s != nil {\n", ctx.valueName)
			tcg.pg.p(forkedTcg.g.string())
			tcg.pg.p("}\n")
			tcg.pg.p(forkedTcg.pg.string())
			return
		}

		forkedTcg := tcg.fork()
		forkedTcg.Gen(&TypConvContext{
			kt:             x.Underlying(),
			keyName:        ctx.keyName,
			vt:             underTpy(ctx.vt),
			valueName:      ctx.valueName,
			assignToken:    ctx.assignToken,
			structTrailing: ctx.structTrailing,
			pValueOnly:     true,
		})
		kv(tcg.g, ctx, ctx.keyName, fmt.Sprintf("%s{\n%s}", tcg.aliaName(x), forkedTcg.g.string()))
		tcg.pg.p(forkedTcg.pg.string())

		return

	case *types.Pointer:
		// Assign same type field
		if types.IdenticalIgnoreTags(ctx.kt, ctx.vt) {
			kv(tcg.g, ctx, ctx.keyName, ctx.valueName)
			return
		}

		// If vt is Pointer, check value is not nil.
		if _, ok := ctx.vt.(*types.Pointer); ok {
			// Prepare assign
			name := ctx.keyName + ctx.valueName
			kv(tcg.g, ctx, ctx.keyName, name)

			forkedTcg := tcg.fork()
			forkedTcg.Gen(&TypConvContext{
				kt:             x.Elem(),
				keyName:        name,
				vt:             underTpy(ctx.vt),
				valueName:      ctx.valueName,
				assignToken:    assignToken_Assign,
				structTrailing: structTrailing_ENTER,
				pValueOnly:     false,
			})
			tcg.pg.p(forkedTcg.pg.string())
			tcg.pg.p("var %s *%s\n", name, tcg.aliaName(x.Elem()))
			tcg.pg.p("if %s != nil {\n", ctx.valueName)
			tcg.pg.p(forkedTcg.g.string())
			tcg.pg.p("}\n")
			return
		}

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
		kv(tcg.g, ctx, ctx.keyName, "&"+forkedTcg.g.string())
		return

	case *types.Slice, *types.Array:
		// TODO[Dokiy] 2022/9/30:
		if types.IdenticalIgnoreTags(ctx.kt, ctx.vt) {
			kv(tcg.g, ctx, ctx.keyName, ctx.valueName)
			return
		}
		return

	case *types.Map:
		if types.IdenticalIgnoreTags(ctx.kt, ctx.vt) {
			kv(tcg.g, ctx, ctx.keyName, ctx.valueName)
			return
		}
		return

	default:
		if types.IdenticalIgnoreTags(ctx.kt, ctx.vt) {
			kv(tcg.g, ctx, ctx.keyName, ctx.valueName)
			return
		}
		return
	}
	// NOTE[Dokiy] 2022/9/29: add_err
	panic("Unsupported type")
}

func kv(g *gener, ctx *TypConvContext, key string, value string) {
	if ctx.pValueOnly {
		g.p("%s", value)
	} else {
		g.p("%s%s %s%s", key, ctx.assignToken, value, ctx.structTrailing)
	}
	return
}
