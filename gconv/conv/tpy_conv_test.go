package conv

import (
	"bytes"
	"context"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

func pkgForMode(path, source string, info *types.Info, mode parser.Mode) (*types.Package, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, source, mode)
	if err != nil {
		return nil, err
	}
	conf := types.Config{Importer: importer.Default()}
	return conf.Check(f.Name.Name, fset, []*ast.File{f}, info)
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

func TestGenTpyConv(t *testing.T) {
	var (
		t_name   = "first test"
		t_path   = "temp"
		t_LIdent = "doData"
		t_RIdent = "daoData"
	)

	t.Run(t_name, func(t *testing.T) {
		gopath, err := filepath.Abs("testdata/tpy_conv/")
		if err != nil {
			t.Fatal(err)
		}
		cfg := &packages.Config{
			Context: context.Background(),
			Mode:    packages.NeedTypes | packages.NeedTypesInfo,
			Env:     append(os.Environ(), "GOPATH="+gopath),
		}
		pkgs, err := packages.Load(cfg, filepath.Join(gopath, t_path))
		if err != nil {
			t.Fatalf("%s: incorrect test src: %s", t_path, err)
		}
		for _, pkg := range pkgs {
			X := pkg.Types.Scope().Lookup(t_LIdent)
			Y := pkg.Types.Scope().Lookup(t_RIdent)
			if X == nil || Y == nil {
				continue
			}

			stmt := GenTpyConv(nil, X.Type(), Y.Type())
			t.Logf("\n%s", strStmts(stmt))
			return
		}
		t.Fatalf("test must declare both %s and %s", t_LIdent, t_RIdent)
	})
}
