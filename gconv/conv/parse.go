package conv

import (
	"go/ast"
	"go/types"
	"strings"
)

// parseImportAlias find all named imports.
func parseImportAlias(syn *ast.File) map[string]string {
	importAlias := make(map[string]string, len(syn.Imports))
	for _, imp := range syn.Imports {
		alias := strings.Trim(imp.Path.Value, "\"")
		// Named import
		if imp.Name != nil {
			importAlias[alias] = strings.Trim(imp.Name.Name, "\"")
			continue
		}

		// Add last name
		index := strings.LastIndex(imp.Path.Value, "/")
		if index < 0 {
			index = 0
		}

		importAlias[alias] = strings.Trim(imp.Path.Value[index+1:], "\"")
	}

	return importAlias
}

// qualifiedIdentObject finds the object for an identifier or a
// qualified identifier, or nil if the object could not be found.
func qualifiedIdentObject(info *types.Info, expr ast.Expr) types.Object {
	switch expr := expr.(type) {
	case *ast.Ident:
		return info.ObjectOf(expr)
	case *ast.SelectorExpr:
		pkgName, ok := expr.X.(*ast.Ident)
		if !ok {
			return nil
		}
		if _, ok := info.ObjectOf(pkgName).(*types.PkgName); !ok {
			return nil
		}
		return info.ObjectOf(expr.Sel)
	default:
		return nil
	}
}

func underPointerTpy(tpy types.Type) types.Type {
	if rtp, ok := tpy.(*types.Pointer); ok {
		tpy = rtp.Elem()
	}
	if tpy.Underlying() == tpy {
		return tpy
	}
	return underPointerTpy(tpy.Underlying())
}
