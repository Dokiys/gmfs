package conv

import (
	"fmt"
	"go/types"
)

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
