package testdata

import (
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/Dokiys/gmfs"
)

func TestValidate2(t *testing.T) {
	f, err := os.Open("template2.go")
	if err != nil {
		t.Fatal(err)
	}

	w := os.Stdout
	rex, _ := regexp.Compile(".*")
	gmfs.RegisterValidateFunc(BindingValidateFunc)
	if err := gmfs.GenMsg(f, w, *rex); err != nil {
		t.Fatal(err)
	}
}
func TestValidate(t *testing.T) {
	f, err := os.Open("validate.go")
	if err != nil {
		t.Fatal(err)
	}

	w := os.Stdout
	rex, _ := regexp.Compile(".*")
	gmfs.RegisterValidateFunc(BindingValidateFunc)
	if err := gmfs.GenMsg(f, w, *rex); err != nil {
		t.Fatal(err)
	}

}

const TagBinding = "binding"
const PrefixRequired = "required"
const RuleOneof = "oneof"

func BindingValidateFunc(field *ast.Field) (v string) {
	if field.Tag == nil {
		return
	}

	tag := reflect.StructTag(field.Tag.Value).Get(TagBinding)

	split := strings.Split(tag, ",")
	if split[0] != PrefixRequired {
		return
	}
	var rule string
	if len(split) >= 2 {
		rule = split[1]
	}

	return genValidate(field.Type, rule)
}

func genValidate(n ast.Node, rule string) (v string) {
	ast.Inspect(n, func(node ast.Node) bool {
		switch node.(type) {
		case *ast.SelectorExpr:
			expr, _ := node.(*ast.SelectorExpr)
			if ident, ok := expr.X.(*ast.Ident); ok && expr.Sel.Name == "Time" && ident.Name == "time" {
				v = "[(validate.rules).timestamp.required = true]"
				return false
			}
			v = genValidate(expr.Sel, rule)
			return false

		case *ast.Ident:
			ident, _ := node.(*ast.Ident)
			switch ident.Name {
			case types.Typ[types.String].Name():
				v = "[(validate.rules).string.min_len = 1]"

			case types.Typ[types.Int].Name():
				if len(rule) <= 0 {
					rule = "gte = 1"
				}
				v = fmt.Sprintf("[(validate.rules).%s.%s]", gmfs.IntType, rule)

			case types.Typ[types.Int32].Name(), types.Typ[types.Int64].Name():
				if len(rule) <= 0 {
					rule = "gte = 1"
				}

				if strings.Contains(rule, RuleOneof) {
					rule = strings.TrimLeft(rule, RuleOneof)
					rule = strings.TrimSpace(rule)
					rule = strings.TrimLeft(rule, "=")

					split := strings.Split(rule, " ")
					rule = fmt.Sprintf(" = {in: [%s]}", strings.Join(split, ","))
				} else {
					rule = "." + rule
				}

				v = fmt.Sprintf("[(validate.rules).%s%s]", ident.Name, rule)
			default:
				v = "[(validate.rules).message.required = true]"
			}

		case *ast.ArrayType:
			v = "[(validate.rules).repeated.min_items = 1]"

		default:
			return true
		}

		return false
	})

	return
}
