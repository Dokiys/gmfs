package gmfs

import (
	"go/ast"
)

type ValidateFunc = func(field *ast.Field) string

var validate = defValidateFunc

func defValidateFunc(_ *ast.Field) string {
	return ""
}

func RegisterValidateFunc(f ValidateFunc) {
	validate = f
}
