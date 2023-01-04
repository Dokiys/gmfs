package conv

import (
	"context"
	"go/ast"
	"go/printer"
	"io"
	"os"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

//type conv struct {
//	buf bytes.Buffer
//	pkg *packages.Package
//}
//
//func newConv(pkg *packages.Package) *conv {
//	return &conv{pkg: pkg}
//}

func Gen(pkgPath string) error {
	ctx := context.Background()
	cfg := &packages.Config{
		Context:    ctx,
		Mode:       packages.NeedName | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		Env:        os.Environ(),
		BuildFlags: []string{"-tags=gconv"},
	}
	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		w := os.Stdout
		gen(w, pkg)
	}
	return nil
}

func gen(w io.Writer, pkg *packages.Package) {
	for i, _ := range pkg.CompiledGoFiles {
		//gcf := cgf[:len(cgf)-3] + "_gconv_gen.go"

		syn := pkg.Syntax[i]
		//node := copyAST(syn)
		//var decls []ast.Decl

		// 替换需要转换到方法内容
		astutil.Apply(syn, func(c *astutil.Cursor) bool {
			switch x := c.Node().(type) {
			case *ast.FuncDecl:
				fnConv, ok := newFnConv(pkg, parseImportAlias(syn), x)
				if !ok {
					return true
				}
				fnConv.replace()
			}

			return true
		}, nil)

		if err := printer.Fprint(w, pkg.Fset, syn); err != nil {
			panic(err)
		}

		//source, err := format.Source(buf.Bytes())
		//if err != nil {
		//	panic(err)
		//}

	}
}