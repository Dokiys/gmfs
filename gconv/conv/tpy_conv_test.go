package conv

import (
	"bytes"
	"context"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestGenTpyConv(t *testing.T) {
	const X, Y = "X", "Y"
	const wantFile = "want"

	var defCtx = &typCtx{
		AssignedIdent: "x",
		AssignIdent:   "y",
	}

	tests := []struct {
		name   string
		typCtx *typCtx
	}{
		{"Basic", defCtx},
		{"Nested", defCtx},
		{"PkgStruct_basic", defCtx},
		{"Pointer_X", defCtx},
		{"Pointer_XY", defCtx},
		{"Pointer_Y", defCtx},
		// TODO[Dokiy] 2022/9/30: arr, slice
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gopath, err := filepath.Abs("testdata/tpy_conv/")
			if err != nil {
				t.Fatalf("%s: get gopath err:%s", tt.name, err)
			}
			cfg := &packages.Config{
				Context: context.Background(),
				Mode:    packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
				Env:     append(os.Environ(), "GOPATH="+gopath),
			}

			pkgs, err := packages.Load(cfg, filepath.Join(gopath, tt.name))
			if err != nil {
				t.Fatalf("%s: incorrect test src: %s", tt.name, err)
			}
			for i, pkg := range pkgs {
				// init pkg alias
				tt.typCtx.PkgAlias = parseImportAlias(pkg.Syntax[i])

				x := pkg.Types.Scope().Lookup(X)
				y := pkg.Types.Scope().Lookup(Y)
				if x == nil || y == nil {
					continue
				}

				got := printConvExpr(x, GenTpyConv(tt.typCtx, x.Type(), y.Type()))
				expected, err := os.ReadFile(filepath.Join(gopath, tt.name, wantFile))
				if err != nil {
					t.Fatalf("%s: read wantFile file err: %s", wantFile, err)
				}
				if got != string(expected) {
					t.Errorf("got:\n%s\nexpected:\n%s\n", got, expected)
				}

				return
			}
			t.Fatalf("test must declare both %s and %s", X, Y)
		})
	}
}

// TODO[Dokiy] 2023/1/4: to be continued!
func printConvExpr(x types.Object, expr []ast.Expr) string {
	var op token.Token
	if _, ok := x.Type().(*types.Pointer); ok {
		op = token.AND
	}

	var stmts []ast.Stmt
	stmts = append(stmts, &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(x.Name())},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.UnaryExpr{
			Op: op,
			X: &ast.CompositeLit{
				Type: &ast.Ident{
					Name: x.Name(),
				},
				Elts: expr,
			},
		}},
	})

	return strStmts(stmts)
}

func strStmts(stmts []ast.Stmt) string {
	// Create a FileSet for node. Since the node does not come
	// from a real source file, fset will be empty.
	fset := token.NewFileSet()
	var buf bytes.Buffer

	err := format.Node(&buf, fset, stmts)
	if err != nil {
		log.Fatal(err)
	}

	return buf.String()
}
