package gmfs

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"regexp"
	"strings"
)

const specTab = "\t"
const specEnter = "\n"
const commentPrefix = "//"

var TypInt = fmt.Sprintf("int%d", 32<<(^uint(0)>>63))

func GenMsg(r io.Reader, w io.Writer, exp regexp.Regexp) error {
	src, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	var messages []string
	fset := token.NewFileSet()
	astf, err := parser.ParseFile(fset, "", string(src), parser.ParseComments)
	if err != nil {
		return err
	}

	cmap := ast.NewCommentMap(fset, astf, astf.Comments)

	var declCmt string
	for _, decl := range astf.Decls {
		d, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		declCmt = fmt.Sprintf("%s", genComment(cmap[d], specEnter))
		for _, spec := range d.Specs {
			switch spec.(type) {
			case *ast.TypeSpec:
				tpy := spec.(*ast.TypeSpec)
				name := tpy.Name.Obj.Name
				if !exp.MatchString(name) {
					continue
				}

				st, ok := tpy.Type.(*ast.StructType)
				if !ok {
					continue
				}

				messages = append(messages, declCmt+specEnter+genMsg(cmap, st, name))
			}
		}
	}

	_, err = fmt.Fprint(w, strings.Join(messages, specEnter))
	return err
}

func genMsg(cmap ast.CommentMap, st *ast.StructType, name string) string {
	var msg = fmt.Sprintf("message %s {"+specEnter, name)

	for i, field := range st.Fields.List {
		msg += fmt.Sprintf("%s"+specEnter, genComment(cmap[field], specTab))

		// unnamed parameters
		if len(field.Names) <= 0 {
			msg += specTab + commentPrefix + " Unsupported field: " + getUnsupportedFieldName(field) + specEnter
			continue
		}

		// gen field
		if isSupported(field.Type) {
			msg += fmt.Sprintf(specTab+"%s %s = %d%s;"+specEnter, genFiledTyp(field.Type), snakeName(field.Names[0].Name), i+1, validate(field))
		} else {
			msg += specTab + commentPrefix + " Unsupported field: " + field.Names[0].Name + specEnter
		}
	}
	msg += "}"

	return msg
}

func genComment(cg []*ast.CommentGroup, spec string) (comment string) {
	last := len(cg) - 1
	if last <= -1 {
		return ""
	}

	for _, c := range cg[last].List {
		comment += spec + c.Text
	}

	return comment
}

func genFiledTyp(expr ast.Expr) (name string) {
	switch x := expr.(type) {
	case *ast.Ident:
		name = getIdentName(x)

	case *ast.SelectorExpr:
		name = getSelectorExprName(x)

	case *ast.StarExpr:
		name = genFiledTyp(x.X)

	case *ast.MapType:
		if k, v := genFiledTyp(x.Key), genFiledTyp(x.Value); (len(k) <= 0 || len(v) <= 0) ||
			strings.HasPrefix(k, "repeated") ||
			strings.HasPrefix(v, "repeated") {
			return ""
		}
		name = fmt.Sprintf("map<%s,%s>", genFiledTyp(x.Key), genFiledTyp(x.Value))

	case *ast.ArrayType:
		if tpyName := genFiledTyp(x.Elt); len(tpyName) <= 0 ||
			strings.HasPrefix(tpyName, "repeated") ||
			strings.HasPrefix(tpyName, "map<") {
			return ""
		}
		name = "repeated" + " " + genFiledTyp(x.Elt)
	}

	return name
}

func isSupported(typ ast.Expr) bool {
	var times int
	ast.Inspect(typ, func(node ast.Node) bool {
		switch node.(type) {
		case *ast.FuncType:
			times += 2
			return false

		case *ast.MapType, *ast.ArrayType:
			if times >= 1 {
				times += 1
				return false
			}
			times += 1
			return true
		}
		return true
	})
	return times < 2
}

func getSelectorExprName(expr *ast.SelectorExpr) (name string) {
	name = expr.Sel.Name
	if expr.Sel.Name == "Time" && genFiledTyp(expr.X) == "time" {
		name = "google.protobuf.Timestamp"
	}
	return
}

func getIdentName(ident *ast.Ident) (name string) {
	switch ident.Name {
	case "int":
		name = TypInt
	case "float64":
		name = "double"
	case "float32":
		name = "float"
	default:
		name = ident.Name
	}
	return
}

func getUnsupportedFieldName(field *ast.Field) string {
	var fieldName string
	var typNames []string

	ast.Inspect(field, func(node ast.Node) bool {
		if _, ok := node.(*ast.StarExpr); ok {
			fieldName = "*"
		}
		if ident, ok := node.(*ast.Ident); ok {
			typNames = append(typNames, ident.Name)
		}
		return true
	})

	fieldName += strings.Join(typNames, ".")
	return fieldName
}

func snakeName(name string) string {
	l := len(name)
	if l <= 0 {
		return ""
	}

	s := make([]byte, 0, l*2)
	s = append(s, name[0])
	for i := 1; i < l; i++ {
		p, c := name[i-1], name[i]
		if p != '_' && c >= 'A' && c <= 'Z' {
			s = append(s, '_')
		}
		s = append(s, c)
	}

	return strings.ToLower(string(s))
}
