package conv

import (
	"context"
	"go/ast"
	"go/printer"
	"go/token"
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

func load(pkgPath string) ([]*packages.Package, error) {
	ctx := context.Background()
	cfg := &packages.Config{
		Context:    ctx,
		Mode:       packages.NeedName | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		Env:        os.Environ(),
		BuildFlags: []string{"-tags=gconv"},
	}
	return packages.Load(cfg, pkgPath)
}

func Gen(pkgPath string) error {
	pkgs, err := load(pkgPath)
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

		astutil.Apply(syn, func(c *astutil.Cursor) bool {
			switch x := c.Node().(type) {
			case *ast.FuncDecl:
				fnConv, ok := newFnConv(pkg, x)
				if !ok {
					return false
				}
				//fnConv.genConvStmt()

				// replace ConvFunc stmt
				replaceFuncStmts(x, fnConv.genConvStmt())
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

func replaceFuncStmts(fn *ast.FuncDecl, stmts []ast.Stmt) {
	astutil.Apply(fn, func(c *astutil.Cursor) bool {
		switch c.Node().(type) {
		case *ast.BlockStmt:
			c.Replace(&ast.BlockStmt{
				Lbrace: token.NoPos,
				List:   stmts,
				Rbrace: token.NoPos,
			})

			return false
		}

		return true
	}, nil)
	return
}
