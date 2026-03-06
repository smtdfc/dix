package parser

import (
	"fmt"
	"go/types"
)

func getPackagePath(t types.Type) string {
	switch t := t.(type) {
	case *types.Named:

		if obj := t.Obj(); obj != nil && obj.Pkg() != nil {
			return obj.Pkg().Path()
		}
	case *types.Pointer:

		return getPackagePath(t.Elem())
	case *types.Slice:

		return getPackagePath(t.Elem())
	case *types.Map:

		return getPackagePath(t.Elem())
	}
	return ""
}

func parseTypeDetails(t types.Type) (string, bool) {
	isPointer := false

	if ptr, ok := t.(*types.Pointer); ok {
		isPointer = true
		t = ptr.Elem()
	}

	typeName := types.TypeString(t, func(p *types.Package) string {
		return ""
	})

	return typeName, isPointer
}

func containsInjectable(comment string) bool {
	return (comment != "" && (fmt.Sprintf("%s", comment) == "@Injectable" ||
		(len(comment) >= 11 && comment[:11] == "@Injectable")))
}
