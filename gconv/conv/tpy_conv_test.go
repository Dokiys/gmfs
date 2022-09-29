package conv

import (
	"bytes"
	"context"
	"go/ast"
	"go/format"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

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

func TestGenTpyConv(t *testing.T) {
	const X, Y = "X", "Y"
	const wantFile = "want"

	tests := []struct {
		name   string
		typCtx *typCtx
	}{
		{"Basic", nil},
		{"Nested", nil},
		{"Nested_Y_pointer", nil},
		{"Nested_X_pointer", nil},
		// TODO[Dokiy] 2022/9/29: pkg.Struct
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gopath, err := filepath.Abs("testdata/tpy_conv/")
			if err != nil {
				t.Fatalf("%s: get gopath err:%s", tt.name, err)
			}
			cfg := &packages.Config{
				Context: context.Background(),
				Mode:    packages.NeedTypes | packages.NeedTypesInfo,
				Env:     append(os.Environ(), "GOPATH="+gopath),
			}
			pkgs, err := packages.Load(cfg, filepath.Join(gopath, tt.name))
			if err != nil {
				t.Fatalf("%s: incorrect test src: %s", tt.name, err)
			}
			for _, pkg := range pkgs {
				x := pkg.Types.Scope().Lookup(X)
				y := pkg.Types.Scope().Lookup(Y)
				if x == nil || y == nil {
					continue
				}

				stmt := GenTpyConv(nil, x.Type(), y.Type())
				got := strStmts(stmt)
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
