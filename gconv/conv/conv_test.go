package conv

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"testing"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

func TestGen(t *testing.T) {
	if err := Gen("gconv/temp"); err != nil {
		t.Fatal(err)
	}
}

// TODO[Dokiy] 2022/8/3: 加载到pacakge中获取类型，并反射创建实例
func TestPackage(t *testing.T) {
	ctx := context.Background()
	cfg := &packages.Config{
		Context:    ctx,
		Mode:       packages.NeedName | packages.NeedCompiledGoFiles | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		Env:        os.Environ(),
		BuildFlags: []string{"-tags=gconv"},
	}
	pkgs, err := packages.Load(cfg, "gconv/temp")
	if err != nil {
		t.Error(err)
	}

	//ast.Inspect(pkgs[0].Syntax[0], inspect)

	//var obj types.Object
	//obj.Type().Underlying().(*types.Struct).NumFields()

	// NOTE[Dokiy] 2022/8/3: 获取入参类型
	//pkgs[0].Syntax[1].Decls[0].Type.Params.List.Type...
	// NOTE[Dokiy] 2022/8/3: 去获取参数
	// pkgs[0].Imports["go_test/my/gencov/data"].Syntax.Tok == Type
	// pkgs[0].Imports["go_test/my/gencov/data"].Syntax.Speces.Type.(StructType).Fields
	defs := pkgs[0].TypesInfo.Defs
	t.Log(defs)
}

func TestAstutil(t *testing.T) {
	src := `
package p

func pred() bool {
  return true
}

func pp(x int) int {
  if x > 2 && pred() {
    return 5
  }

  var b = pred()
  if b {
    return 6
  }
  return 0
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		log.Fatal(err)
	}

	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		switch x := n.(type) {
		case *ast.CallExpr:
			id, ok := x.Fun.(*ast.Ident)
			if ok {
				if id.Name == "pred" {
					c.Replace(&ast.UnaryExpr{
						Op: token.NOT,
						X:  x,
					})
				}
			}
		}

		return true
	})

	fmt.Println("Modified AST:")
	printer.Fprint(os.Stdout, fset, file)
}
