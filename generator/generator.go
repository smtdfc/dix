package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"

	"github.com/smtdfc/dix/parser"
)

type Generator struct{}

const generatedBuildHeader = "//go:build !dix\n// +build !dix\n\n"

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) GenerateImportStmt(scope *Scope) (*ast.GenDecl, error) {
	specs := []ast.Spec{}

	for path, ident := range scope.Imports {
		specs = append(specs, &ast.ImportSpec{
			Name: ident,
			Path: &ast.BasicLit{
				Value: fmt.Sprintf("%q", path),
				Kind:  token.STRING,
			},
		})
	}

	return &ast.GenDecl{
		Tok:    token.IMPORT,
		Lparen: token.Pos(1),
		Specs:  specs,
	}, nil
}

func (g *Generator) GenerateDep(dep *parser.Dependency, scope *Scope) (ast.Expr, error) {
	return g.GenerateDepWithMap(dep, scope, nil)
}

func (g *Generator) GenerateDepWithMap(dep *parser.Dependency, scope *Scope, providerMap map[string]*parser.Provider) (ast.Expr, error) {
	if dep.IsSingleton {
		if providerMap == nil {
			return nil, NewGenerateError(
				ErrorCodeGeneration,
				"provider map is required to generate singleton dependency",
				"",
				dep.Type.Signature(),
				nil,
			)
		}

		provider, ok := providerMap[dep.Type.Signature()]
		if !ok {
			return nil, NewGenerateError(
				ErrorDependencyResolve,
				"singleton dependency provider not found",
				"",
				dep.Type.Signature(),
				nil,
			)
		}

		providerCall, err := g.GenerateCallProviderWithMap(provider, scope, providerMap)
		if err != nil {
			return nil, err
		}

		diPkg := scope.Import("github.com/smtdfc/dix/di")
		return &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   diPkg,
				Sel: ast.NewIdent("NewSingleton"),
			},
			Args: []ast.Expr{providerCall},
		}, nil
	}

	ident, ok := scope.Names[dep.Type.Signature()]
	if !ok {
		return nil, NewGenerateError(
			ErrorDependencyResolve,
			"dependency is not available in generated container scope",
			"",
			dep.Type.Signature(),
			nil,
		)
	}

	return ident, nil
}

func (g *Generator) GenerateCallProvider(provider *parser.Provider, scope *Scope) (ast.Expr, error) {
	return g.GenerateCallProviderWithMap(provider, scope, nil)
}

func (g *Generator) GenerateCallProviderWithMap(provider *parser.Provider, scope *Scope, providerMap map[string]*parser.Provider) (ast.Expr, error) {
	args := []ast.Expr{}

	for _, dep := range provider.Deps {
		argExpr, err := g.GenerateDepWithMap(dep, scope, providerMap)
		if err != nil {
			return nil, err
		}
		args = append(args, argExpr)
	}

	pkgAlias := scope.Import(provider.PackagePath)
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   pkgAlias,
			Sel: ast.NewIdent(provider.Name),
		},
		Args: args,
	}, nil
}

func (g *Generator) GenerateCreateObjectStmt(ident *ast.Ident, provider *parser.Provider, scope *Scope, providerMap map[string]*parser.Provider) (ast.Stmt, error) {
	callExpr, err := g.GenerateCallProviderWithMap(provider, scope, providerMap)
	if err != nil {
		return nil, err
	}

	return &ast.AssignStmt{
		Lhs: []ast.Expr{ident},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{callExpr},
	}, nil
}

func (g *Generator) Generate(metadata *parser.Metadata) (string, error) {
	if metadata.Root == nil {
		return "", NewGenerateError(ErrorValidation, "cannot find @Root provider", "", "", nil)
	}

	scope := NewScope()
	providerMap := make(map[string]*parser.Provider)
	for _, c := range metadata.Providers {
		providerMap[c.Return.Type.Signature()] = c
	}

	graph, err := BuildGraph(metadata.Root, providerMap)
	if err != nil {
		return "", err
	}

	// Track provider return types required as regular (non-singleton) dependencies.
	nonSingletonTypes := make(map[string]bool)
	for _, provider := range metadata.Providers {
		for _, dep := range provider.Deps {
			if !dep.IsSingleton {
				nonSingletonTypes[dep.Type.Signature()] = true
			}
		}
	}
	if metadata.Root != nil {
		for _, dep := range metadata.Root.Deps {
			if !dep.IsSingleton {
				nonSingletonTypes[dep.Type.Signature()] = true
			}
		}
	}

	sorted, err := graph.Sort()
	if err != nil {
		return "", err
	}

	fset := token.NewFileSet()
	stmts := []ast.Stmt{}

	for _, provider := range sorted {
		isRoot := provider.Return.Type.Signature() == metadata.Root.Return.Type.Signature()
		if !isRoot && !nonSingletonTypes[provider.Return.Type.Signature()] {
			continue
		}

		id := scope.UniqueIdent(provider.Return.Type.Name)
		stmt, err := g.GenerateCreateObjectStmt(id, provider, scope, providerMap)
		if err != nil {
			return "", err
		}

		scope.Names[provider.Return.Type.Signature()] = id
		stmts = append(stmts, stmt)
	}

	lastComp := sorted[len(sorted)-1]
	finalID, ok := scope.Names[lastComp.Return.Type.Signature()]
	if !ok {
		return "", NewGenerateError(ErrorCodeGeneration, "failed to resolve root provider identifier", lastComp.Name, "", nil)
	}

	var finalExpr ast.Expr = finalID

	stmts = append(stmts, &ast.ReturnStmt{
		Results: []ast.Expr{finalExpr},
	})

	fn := &ast.FuncDecl{
		Name: ast.NewIdent("Root"),
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{Type: g.TypeToASTExpr(lastComp.Return.Type, scope)},
				},
			},
		},
		Body: &ast.BlockStmt{List: stmts},
	}

	importDecl, err := g.GenerateImportStmt(scope)
	if err != nil {
		return "", err
	}

	file := &ast.File{
		Name:  ast.NewIdent("generated"),
		Decls: []ast.Decl{importDecl, fn},
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return "", err
	}

	return generatedBuildHeader + buf.String(), nil
}

func (g *Generator) TypeToASTExpr(t *parser.TypeInfo, scope *Scope) ast.Expr {
	var expr ast.Expr
	if t.Pkg != "" {
		alias := scope.Import(t.Pkg)
		expr = &ast.SelectorExpr{
			X:   alias,
			Sel: ast.NewIdent(t.Name),
		}
	} else {
		expr = ast.NewIdent(t.Name)
	}

	if t.IsPointer {
		expr = &ast.StarExpr{X: expr}
	}
	return expr
}
