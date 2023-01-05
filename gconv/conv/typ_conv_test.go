package conv

import (
	"context"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestGenTpyConv(t *testing.T) {
	const X, Y = "X", "Y"
	const wantFile = "want"

	tests := []struct {
		name string
	}{
		{"Basic"},
		// TODO[Dokiy] 2023/1/5: to be continued!
		{"Nested"},
		{"PkgStruct_basic"},
		{"Pointer_X"},
		{"Pointer_XY"},
		{"Pointer_Y"},
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
				x := pkg.Types.Scope().Lookup(X)
				y := pkg.Types.Scope().Lookup(Y)
				if x == nil || y == nil {
					continue
				}

				tcg := &TypConvGen{
					Ctx:      NewTypCtx(x.Name(), y.Name()),
					g:        newGener(""),
					pkgAlias: parseImportAlias(pkg.Syntax[i]),
					ignore:   nil,
					kt:       x.Type(),
					vt:       y.Type(),
				}
				tcg.gen()

				got, err := format.Source([]byte(tcg.g.string()))
				if err != nil {
					t.Fatalf("%s: format genSrc err: %s", wantFile, err)
				}
				expected, err := os.ReadFile(filepath.Join(gopath, tt.name, wantFile))
				if err != nil {
					t.Fatalf("%s: read wantFile file err: %s", wantFile, err)
				}
				if strings.Compare(string(got), string(expected)) != 0 {
					t.Errorf("got:\n%s\nexpected:\n%s\n", got, expected)
				}

				return
			}
			t.Fatalf("test must declare both %s and %s", X, Y)
		})
	}
}
