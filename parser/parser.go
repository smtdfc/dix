package parser

import (
	"fmt"
	"go/ast"
	"go/types"

	"github.com/fatih/color"
	"golang.org/x/tools/go/packages"
)

type Parser struct{}

func isDixSingletonNamed(named *types.Named) bool {
	if named == nil {
		return false
	}

	origin := named.Origin()
	if origin == nil || origin.Obj() == nil || origin.Obj().Pkg() == nil {
		return false
	}

	return origin.Obj().Name() == "Singleton" &&
		origin.Obj().Pkg().Path() == "github.com/smtdfc/dix/di"
}

func (p *Parser) ParseProvider(pkg *packages.Package, file *ast.File, fn *ast.FuncDecl) (*Provider, error) {
	c := &Provider{
		Name:        fn.Name.Name,
		File:        pkg.Fset.Position(file.Package).Filename,
		PackagePath: pkg.PkgPath,
		PackageName: pkg.Name,
	}

	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			_, isPointerAtAST := field.Type.(*ast.StarExpr)

			tv := pkg.TypesInfo.TypeOf(field.Type)
			if tv == nil {
				continue
			}

			for _, name := range field.Names {
				var depType *types.Type
				isSingleton := false

				var singletonNamed *types.Named
				switch t := tv.(type) {
				case *types.Named:
					if isDixSingletonNamed(t) {
						singletonNamed = t
					}
				case *types.Pointer:
					if n, ok := t.Elem().(*types.Named); ok && isDixSingletonNamed(n) {
						return nil, NewValidationError(
							"singleton dependency must be `di.Singleton[T]`, not `*di.Singleton[T]`",
							fn.Name.Name,
							name.Name,
							c.File,
						)
					}
				}

				if singletonNamed != nil {
					if isPointerAtAST {
						return nil, NewValidationError(
							"singleton dependency must be `di.Singleton[T]`, not pointer form",
							fn.Name.Name,
							name.Name,
							c.File,
						)
					}

					targs := singletonNamed.TypeArgs()
					if targs != nil && targs.Len() > 0 {
						innerType := targs.At(0)
						depType = &innerType
						isSingleton = true
					}
				}

				if depType == nil {
					depType = &tv
				}

				typeName, isPtr := parseTypeDetails(*depType)
				pkgPath := getPackagePath(*depType)

				c.Deps = append(c.Deps, &Dependency{
					Name: name.Name,
					Type: &TypeInfo{
						Name:      typeName,
						Pkg:       pkgPath,
						IsPointer: isPtr,
					},
					IsSingleton: isSingleton,
				})
			}
		}
	}

	if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
		return nil, NewValidationError(
			"provider function must return exactly one value",
			fn.Name.Name,
			"",
			c.File,
		)
	}

	if len(fn.Type.Results.List) > 1 || (len(fn.Type.Results.List) == 1 && len(fn.Type.Results.List[0].Names) > 1) {
		return nil, NewValidationError(
			"provider function must return exactly one value",
			fn.Name.Name,
			"",
			c.File,
		)
	}

	for _, field := range fn.Type.Results.List {
		rtv := pkg.TypesInfo.TypeOf(field.Type)
		rName, rIsPtr := parseTypeDetails(rtv)
		rPkg := getPackagePath(rtv)

		c.Return = &ReturnValue{
			Type: &TypeInfo{
				Name:      rName,
				Pkg:       rPkg,
				IsPointer: rIsPtr,
			},
		}
	}

	return c, nil
}

func (p *Parser) Parse(dir string) (*Metadata, error) {

	cfg := &packages.Config{
		Dir:        dir,
		BuildFlags: []string{"-tags=dix"},
		Mode:       packages.NeedName | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedImports,
	}

	metadata := new(Metadata)

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, err
	}

	var parseErr error
	for _, pkg := range pkgs {

		if len(pkg.Errors) > 0 {
			return nil, NewPackageLoadError(pkg.Errors[0])
		}

		for _, file := range pkg.Syntax {

			fileName := pkg.Fset.Position(file.Package).Filename

			fmt.Printf("\033[32m[Scan]\033[0m File: %s ... ", fileName)

			ast.Inspect(file, func(n ast.Node) bool {
				fn, ok := n.(*ast.FuncDecl)
				if !ok || fn.Doc == nil {
					return true
				}

				if containsInjectableAnnotation(fn.Doc.Text()) {
					m, err := p.ParseProvider(pkg, file, fn)
					if err != nil {
						parseErr = err
						return false
					}

					if containsRootAnnotation(fn.Doc.Text()) {
						metadata.Root = m
					} else {
						metadata.Providers = append(metadata.Providers, m)
					}

					if containsDisableAnnotation(fn.Doc.Text()) {
						m.IsDisable = true
					}
				}
				return true
			})
			if parseErr != nil {
				// Close the in-progress scan line before printing fatal error output.
				fmt.Println()
				return nil, parseErr
			}

			color.New(color.FgGreen).Printf("OK\n")

		}

	}

	return metadata, nil
}
func NewParser() *Parser {
	return &Parser{}
}
