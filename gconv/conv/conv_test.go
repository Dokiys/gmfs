package conv

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"log"
	"os"
	"sort"
	"testing"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
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

func ExampleMap() {
	const source = `package P

var X []string
var Y []string

const p, q = 1.0, 2.0

func f(offset int32) (value byte, ok bool)
func g(rune) (uint8, bool)
`

	// Parse and type-check the package.
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "P.go", source, 0)
	if err != nil {
		panic(err)
	}
	pkg, err := new(types.Config).Check("P", fset, []*ast.File{f}, nil)
	if err != nil {
		panic(err)
	}

	scope := pkg.Scope()

	// Group names of package-level objects by their type.
	var namesByType typeutil.Map // value is []string
	for _, name := range scope.Names() {
		T := scope.Lookup(name).Type()

		names, _ := namesByType.At(T).([]string)
		names = append(names, name)
		namesByType.Set(T, names)
	}

	// Format, sort, and print the map entries.
	var lines []string
	namesByType.Iterate(func(T types.Type, names interface{}) {
		lines = append(lines, fmt.Sprintf("%s   %s", names, T))
	})
	sort.Strings(lines)
	for _, line := range lines {
		fmt.Println(line)
	}

	// Output:
	// [X Y]   []string
	// [f g]   func(offset int32) (value byte, ok bool)
	// [p q]   untyped float
}

// Issue 16464
func TestAlignofNaclSlice(t *testing.T) {
	const src = `
package main

var s struct {
	x *int
	y []byte
}
`
	ts := findStructType(t, src)
	sizes := &types.StdSizes{WordSize: 4, MaxAlign: 8}
	var fields []*types.Var
	// Make a copy manually :(
	for i := 0; i < ts.NumFields(); i++ {
		fields = append(fields, ts.Field(i))
	}
	offsets := sizes.Offsetsof(fields)
	if offsets[0] != 0 || offsets[1] != 4 {
		t.Errorf("OffsetsOf(%v) = %v want %v", ts, offsets, []int{0, 4})
	}
}

// findStructType typechecks src and returns the first struct type encountered.
func findStructType(t *testing.T, src string) *types.Struct {
	return findStructTypeConfig(t, src, &types.Config{})
}

func findStructTypeConfig(t *testing.T, src string, conf *types.Config) *types.Struct {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "x.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	info := types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}
	_, err = conf.Check("x", fset, []*ast.File{f}, &info)
	if err != nil {
		t.Fatal(err)
	}
	for _, tv := range info.Types {
		if ts, ok := tv.Type.(*types.Struct); ok {
			return ts
		}
	}
	t.Fatalf("failed to find a struct type in src:\n%s\n", src)
	return nil
}
