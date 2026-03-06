package parser

import (
	"fmt"
	"go/ast"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

type Parser struct{}

func (p *Parser) ParseConstructorFn(pkg *packages.Package, fn *ast.FuncDecl) (*Constructor, error) {
	c := new(Constructor)
	c.Name = fn.Name.Name

	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			for _, name := range field.Names {
				tv := pkg.TypesInfo.TypeOf(field.Type)
				pkgPath := getPackagePath(tv)
				typeName, isPointer := parseTypeDetails(tv)
				c.Deps = append(c.Deps, &Dependency{
					Name: name.Name,
					Type: &TypeInfo{
						Name:      typeName,
						Pkg:       pkgPath,
						IsPointer: isPointer,
					},
				})

			}
		}
	}

	if fn.Type.Results != nil {

		if len(fn.Type.Results.List) > 1 {
			return nil, NewParseError(fmt.Sprintf("Constructor function %s is not return more than one value", fn.Name.Name))
		}

		for _, field := range fn.Type.Results.List {
			tv := pkg.TypesInfo.TypeOf(field.Type)

			typeName, isPtr := parseTypeDetails(tv)
			pkgPath := getPackagePath(tv)

			c.Return = &ReturnValue{
				Type: &TypeInfo{
					Name:      typeName,
					Pkg:       pkgPath,
					IsPointer: isPtr,
				},
			}

		}
	}
	return c, nil
}

func (p *Parser) Parse(fileName, code string) (*Metadata, error) {

	absPath, err := filepath.Abs(fileName)
	if err != nil {
		return nil, err
	}

	cfg := &packages.Config{

		Mode: packages.NeedName | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedImports,
		Overlay: map[string][]byte{
			absPath: []byte(code),
		},
	}

	metadata := new(Metadata)

	pkgs, err := packages.Load(cfg, absPath)
	if err != nil {
		return nil, err
	}

	var parseErr error
	for _, pkg := range pkgs {

		if len(pkg.Errors) > 0 {
			return nil, fmt.Errorf("package load error: %v", pkg.Errors[0])
		}

		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				fn, ok := n.(*ast.FuncDecl)
				if !ok || fn.Doc == nil {
					return true
				}

				if containsInjectable(fn.Doc.Text()) {
					m, err := p.ParseConstructorFn(pkg, fn)
					if err != nil {
						parseErr = err
						return false
					}
					metadata.Constructors = append(metadata.Constructors, m)
				}
				return true
			})
			if parseErr != nil {
				return nil, parseErr
			}
		}
	}

	return metadata, nil
}
func NewParser() *Parser {
	return &Parser{}
}
