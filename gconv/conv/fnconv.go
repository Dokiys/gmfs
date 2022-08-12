package conv

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

var emptyIgnoreMap = make(ignoreMap)

type ignoreMap map[string]struct{}

func (i ignoreMap) pickField(stmts []ast.Stmt, xname string) {
	for _, stmt := range stmts {
		switch x := stmt.(type) {
		case *ast.AssignStmt:
			for _, lh := range x.Lhs {
				// check is result assigned
				se, ok := lh.(*ast.SelectorExpr)
				if !ok {
					continue
				}

				ident, ok := se.X.(*ast.Ident)
				if ident.Name != xname {
					continue
				}

				i[se.Sel.Name] = struct{}{}
			}

		default:

		}
	}
}
func (i ignoreMap) exist(name string) bool {
	_, ok := i[name]
	return ok
}

// TODO[Dokiy] 2022/8/12: notePosition reference wire[https://github.com/google/wire/blob/d07cde0df9c5edd46e05e21d29eb315e0b452cbc/internal/wire/errors.go#L60]
type fnConv struct {
	pkg      *packages.Package
	fd       *ast.FuncDecl
	param    *ast.Field
	result   *ast.Field
	ignore   ignoreMap
	convStmt []ast.Stmt
}

// newFnConv
func newFnConv(pkg *packages.Package, fd *ast.FuncDecl) (*fnConv, bool) {
	// check is handle func
	if fd.Recv != nil {
		return nil, false
	}
	if len(fd.Type.Params.List) != 1 || len(fd.Type.Results.List) != 1 {
		return nil, false
	}

	return &fnConv{
		pkg:    pkg,
		fd:     fd,
		ignore: make(ignoreMap),
		param:  fd.Type.Params.List[0],
		result: fd.Type.Results.List[0],
	}, true
}

func (f *fnConv) resultName() string {
	if f.result.Names != nil {
		return f.result.Names[0].Name
	}

	var name string
	ast.Inspect(f.result, func(node ast.Node) bool {
		ident, ok := node.(*ast.Ident)
		if ok {
			name = strings.ToLower(ident.Name)
		}
		return true
	})

	return name
}

func (f *fnConv) genConvStmt() []ast.Stmt {
	f.loadPanicStmt()
	f.convFields()
	return f.stmt()
}

// loadStmt load last panic function stmt.
func (f *fnConv) loadPanicStmt() {
	var stmts []ast.Stmt
	for i, stmt := range f.fd.Body.List {
		switch stmt := stmt.(type) {
		case *ast.EmptyStmt:
			// Do nothing.
		case *ast.ReturnStmt:
			// Do nothing.
		case *ast.ExprStmt:
			call, ok := stmt.X.(*ast.CallExpr)
			if !ok {
				continue
			}

			convObj := qualifiedIdentObject(f.pkg.TypesInfo, call.Fun)
			if convObj != types.Universe.Lookup("panic") || i+1 != len(f.fd.Body.List) {
				continue
			}

			// Handle last panic only.
			for _, arg := range call.Args {
				fl, ok := arg.(*ast.FuncLit)
				if !ok {
					continue
				}

				stmts = append(stmts, fl.Body.List...)
			}
		}
	}

	f.convStmt = stmts
	f.ignore.pickField(stmts, f.resultName())
	return
}

func (f *fnConv) convFields() {
	rt := f.pkg.TypesInfo.TypeOf(f.result.Type)
	pt := f.pkg.TypesInfo.TypeOf(f.param.Type)
	pVar, rVar := getFields(pt, emptyIgnoreMap), getFields(rt, f.ignore)

	// TODO[Dokiy] 2022/8/12: 比较field, 生成 ast.stmt
	// NOTE[Dokiy] 2022/8/12: 处理dao.Item
	// NOTE[Dokiy] 2022/8/12: 处理[]*dao.Item
	fmt.Println(pVar, rVar)
}

func getFields(tpy types.Type, ignore ignoreMap) []*types.Var {
	for tpy.Underlying() != tpy {
		tpy = tpy.Underlying()
	}

	var fields []*types.Var
	for {
		switch x := tpy.(type) {
		case *types.Struct:
			for i := 0; i < x.NumFields(); i++ {
				field := x.Field(i)

				if ignore.exist(field.Name()) || !field.Exported() {
					continue
				}

				fields = append(fields, field)
			}
			return fields
		// TODO[Dokiy] 2022/8/12:
		//case *types.Slice:
		//tpy = x.Elem().Underlying()
		//continue

		case *types.Pointer:
			tpy = x.Elem().Underlying()

		default:
			// TODO[Dokiy] 2022/8/12: notePosition
			panic("Unsupported params")
		}
	}
}

func (f *fnConv) stmt() []ast.Stmt {
	//return f.convStmt
	return f.convStmt
}
