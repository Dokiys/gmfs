package conv

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ast/astutil"
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
	pkg       *packages.Package
	fd        *ast.FuncDecl
	param     *ast.Field
	result    *ast.Field
	ignore    ignoreMap
	convStmt  []ast.Stmt
	panicStmt []ast.Stmt
}

// newFnConv
func newFnConv(pkg *packages.Package, fd *ast.FuncDecl) (*fnConv, bool) {
	// check is handle func
	if fd.Recv != nil {
		return nil, false
	}
	// Make sure one param which name is not '_' and one result
	if len(fd.Type.Params.List) <= 0 || fd.Type.Params.List[0].Names[0].Name == "_" || len(fd.Type.Results.List) != 1 {
		// NOTE[Dokiy] 2022/8/16: notePosition
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
			name = "gconv" + ident.Name
			f.result.Names = []*ast.Ident{ast.NewIdent(name)}
		}
		return true
	})

	return name
}

func (f *fnConv) paramName() string {
	if f.param.Names != nil {
		return f.param.Names[0].Name
	}

	var name string
	ast.Inspect(f.param, func(node ast.Node) bool {
		ident, ok := node.(*ast.Ident)
		if ok {
			name = "gconv" + ident.Name
			f.result.Names = []*ast.Ident{ast.NewIdent(name)}
		}
		return true
	})

	return name
}

func (f *fnConv) loadResultName() {
	if f.result.Names == nil {
		ast.Inspect(f.result, func(node ast.Node) bool {
			ident, ok := node.(*ast.Ident)
			if !ok {
				return true
			}
			f.result.Names = []*ast.Ident{ast.NewIdent("gconv" + ident.Name)}
			return false
		})
	}
}

func (f *fnConv) replaceFunc() {
	// Make sure result name exist
	f.loadResultName()
	f.loadPanicStmt()

	// Conv fields
	rTpy := f.pkg.TypesInfo.TypeOf(f.result.Type)
	pTpy := f.pkg.TypesInfo.TypeOf(f.param.Type)
	pMap, rFields := getFieldsMap(pTpy, emptyIgnoreMap), getFields(rTpy, f.ignore)

	// TODO[Dokiy] 2022/8/16: 2. 如果是数组或者切片，添加外部for语句, 注意assign名称
	// for i = 0; i<len(paramName); i ++ {}
	var stmt []ast.Stmt
	for _, rf := range rFields {
		pf, ok := pMap[rf.Name()]
		if !ok {
			continue
		}
		if varStmt, ok := genVarConv(rf, pf, f.resultName(), f.paramName()); ok {
			stmt = append(stmt, varStmt...)
		}
	}
	f.convStmt = stmt

	// Replace func content
	astutil.Apply(f.fd, func(c *astutil.Cursor) bool {
		switch c.Node().(type) {
		case *ast.BlockStmt:
			c.Replace(&ast.BlockStmt{
				Lbrace: token.NoPos,
				List:   f.content(),
				Rbrace: token.NoPos,
			})

			return false
		}

		return true
	}, nil)
	return
}

// loadStmt load last panic content.
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

	f.panicStmt = stmts
	f.ignore.pickField(stmts, f.resultName())
	return
}

func getFields(tpy types.Type, ignore ignoreMap) []*types.Var {
	var fields []*types.Var
	for {
		switch x := underPointerTpy(tpy).(type) {
		case *types.Struct:
			for i := 0; i < x.NumFields(); i++ {
				field := x.Field(i)

				if ignore.exist(field.Name()) || !field.Exported() {
					continue
				}

				//fields[field.Name()] = field
				fields = append(fields, field)
			}
			return fields
		case *types.Slice:
			tpy = x.Elem()

		case *types.Array:
			tpy = x.Elem()

		default:
			// TODO[Dokiy] 2022/8/12: notePosition
			panic("Unsupported params")
		}
	}
}

func getFieldsMap(tpy types.Type, ignore ignoreMap) map[string]*types.Var {
	var fields = make(map[string]*types.Var)
	for {
		switch x := underPointerTpy(tpy).(type) {
		case *types.Struct:
			for i := 0; i < x.NumFields(); i++ {
				field := x.Field(i)

				if ignore.exist(field.Name()) || !field.Exported() {
					continue
				}

				fields[field.Name()] = field
				//fields = append(fields, field)
			}
			return fields
		// TODO[Dokiy] 2022/8/12:
		//case *types.Slice:
		//tpy = x.Elem().Underlying()
		//continue

		default:
			// TODO[Dokiy] 2022/8/12: notePosition
			panic("Unsupported params")
		}
	}
}

func (f *fnConv) content() (stmt []ast.Stmt) {
	stmt = append(f.convStmt, f.panicStmt...)
	return append(stmt, &ast.ReturnStmt{})
}
