package parser

import (
	"go/types"
	"regexp"
)

var rootRegex = regexp.MustCompile(`(?m)^@Root\s*$`)
var singletonRegex = regexp.MustCompile(`(?m)^@Singleton\s*$`)
var disableRegex = regexp.MustCompile(`(?m)^@Disable\s*$`)
var injectableRegex = regexp.MustCompile(`(?m)^@Injectable\s*$`)

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

func containsInjectableAnnotation(comment string) bool {
	return injectableRegex.MatchString(comment)
}

func containsRootAnnotation(comment string) bool {
	return rootRegex.MatchString(comment)
}

func containsSingletonAnnotation(comment string) bool {
	return singletonRegex.MatchString(comment)
}

func containsDisableAnnotation(comment string) bool {
	return disableRegex.MatchString(comment)
}
