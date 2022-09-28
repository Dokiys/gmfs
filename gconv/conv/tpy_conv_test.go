package conv

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// pkgFor parses and type checks the package specified by path and source,
// populating info if provided.
//
// If source begins with "package generic_" and type parameters are enabled,
// generic code is permitted.
func pkgFor(path, source string, info *types.Info) (*types.Package, error) {
	mode := modeForSource(source)
	return pkgForMode(path, source, info, mode)
}

func pkgForMode(path, source string, info *types.Info, mode parser.Mode) (*types.Package, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, source, mode)
	if err != nil {
		return nil, err
	}
	conf := types.Config{Importer: importer.Default()}
	return conf.Check(f.Name.Name, fset, []*ast.File{f}, info)
}

// genericPkg is a prefix for packages that should be type checked with
// generics.
const genericPkg = "package generic_"

func modeForSource(src string) parser.Mode {
	if !strings.HasPrefix(src, genericPkg) {
		return 1 << 30
	}
	return 0
}

func TestGenTpyConv(t *testing.T) {
	const root = "./testdata/tpy_conv"
	path := filepath.Join(root, "temp")
	src, err := os.ReadFile(filepath.Join(path, "template.go"))
	if err != nil {
		t.Fatalf("%s: incorrect test src: %s", path, err)
	}

	// TODO[Dokiy] 2022/9/28:
	pkg, err := pkgFor("../../gconv", string(src), nil)
	if err != nil {
		t.Fatalf("%s: incorrect test case: %s", path, err)
	}

	X := pkg.Scope().Lookup("X")
	Y := pkg.Scope().Lookup("Y")
	if X == nil || Y == nil {
		t.Fatalf("test must declare both X and Y")
	}

	stmt := GenTpyConv(nil, X.Type(), Y.Type())
	t.Log(stmt)
}
