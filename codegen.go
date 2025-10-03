package dix

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// BuildOrder performs a topological sort using Kahn’s algorithm
// and returns the items in dependency order.
func BuildOrder(container map[string]*Factory) ([]string, error) {
	indegree := map[string]int{}
	graph := map[string][]string{}

	// Build indegree + adjacency
	for alias, f := range container {
		if _, ok := indegree[alias]; !ok {
			indegree[alias] = 0
		}
		for _, dep := range f.Deps {
			if container[dep.Name].Final {
				return nil, fmt.Errorf("final item %s cannot be a dependency of %s", dep.Name, alias)
			}

			if container[dep.Name].Disable {
				return nil, fmt.Errorf("disable item %s cannot be a dependency of %s", dep.Name, alias)
			}

			graph[dep.Name] = append(graph[dep.Name], alias)
			indegree[alias]++
		}
	}

	// Queue nodes with indegree 0
	queue := []string{}
	for n, deg := range indegree {
		if deg == 0 {
			queue = append(queue, n)
		}
	}

	order := []string{}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		order = append(order, node)

		for _, next := range graph[node] {
			indegree[next]--
			if indegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	if len(order) != len(container) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	normal := []string{}
	finals := []string{}
	for _, alias := range order {
		if container[alias].Final {
			finals = append(finals, alias)
		} else {
			normal = append(normal, alias)
		}
	}

	return append(normal, finals...), nil
}

type GenContext struct {
	ImportID  map[string]string // module path -> alias
	Container map[string]string // alias -> var id
	Counter   int
}

// GenUID generates a unique variable name.
func (c *GenContext) GenUID() string {
	c.Counter++
	return "id_" + strconv.Itoa(c.Counter)
}

// ResolveImport ensures the import path has an alias and
// returns the alias for the given module path.
func (c *GenContext) ResolveImport(module string) string {
	if _, ok := c.ImportID[module]; !ok {
		c.ImportID[module] = c.GenUID()
	}
	return c.ImportID[module]
}

// generateDepExpr generates an AST expression for a dependency.
func generateDepExpr(ctx *GenContext, dep *Dependency, config *DIConfig) ast.Expr {
	if dep.Standalone {
		// create new instance
		factory := config.Container[dep.Name]
		modAlias := ctx.ResolveImport(factory.Module)
		args := []ast.Expr{}
		for _, sub := range factory.Deps {
			args = append(args, generateDepExpr(ctx, sub, config))
		}
		return &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(modAlias),
				Sel: ast.NewIdent(factory.Function),
			},
			Args: args,
		}
	}
	//reuse container
	return ast.NewIdent(ctx.Container[dep.Name])
}

// sanitizeModulePath maps an absolute filesystem path to a proper import path
// relative to the project root, prefixed with moduleName.
func sanitizeModulePath(absPath, root, moduleName string) string {
	absPath = filepath.ToSlash(absPath)
	root = filepath.ToSlash(root)

	rel, err := filepath.Rel(root, absPath)
	if err != nil {
		// fallback: return moduleName only
		return moduleName
	}
	rel = filepath.ToSlash(rel)
	rel = strings.TrimPrefix(rel, "./")
	if rel == "." {
		return moduleName
	}
	return moduleName + "/" + rel
}

// GenerateCode generates Go source code that wires all items
// in the dependency injection container.
func GenerateCode(root string, moduleName string, config *DIConfig) (string, error) {
	ctx := &GenContext{
		ImportID:  make(map[string]string),
		Container: make(map[string]string),
		Counter:   0,
	}

	order, err := BuildOrder(config.Container)
	if err != nil {
		return "", err
	}

	// normalize module paths for all factories
	for _, f := range config.Container {
		f.Module = sanitizeModulePath(f.Module, root, moduleName)
	}

	file := &ast.File{
		Name:  ast.NewIdent("dix"),
		Decls: []ast.Decl{},
	}

	stmts := []ast.Stmt{}
	var depIdents []ast.Expr

	for _, item := range order {
		factory := config.Container[item]
		id := factory.Alias
		ctx.Container[item] = id

		args := []ast.Expr{}
		for _, dep := range factory.Deps {
			args = append(args, generateDepExpr(ctx, dep, config))
		}

		modAlias := ctx.ResolveImport(factory.Module)

		assign := &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(id)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent(modAlias),
						Sel: ast.NewIdent(factory.Function),
					},
					Args: args,
				},
			},
		}

		directive := &ast.CommentGroup{
			List: []*ast.Comment{
				{
					Text: fmt.Sprintf("\n//line %s %s", factory.File, factory.Pos),
				},
			},
		}

		stmt := &ast.DeclStmt{
			Decl: &ast.GenDecl{
				Doc: directive,
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names:  []*ast.Ident{ast.NewIdent(id)},
						Values: []ast.Expr{assign.Rhs[0]},
					},
				},
			},
		}

		stmts = append(stmts, stmt)
		depIdents = append(depIdents, ast.NewIdent(id))
	}

	if len(depIdents) > 0 {
		call := &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent("dix"),
				Sel: ast.NewIdent("Mark"),
			},
			Args: depIdents,
		}
		stmts = append(stmts, &ast.ExprStmt{X: call})
	}

	rootFn := &ast.FuncDecl{
		Name: ast.NewIdent("Root"),
		Type: &ast.FuncType{
			Params:  &ast.FieldList{},
			Results: nil,
		},
		Body: &ast.BlockStmt{
			List: stmts,
		},
	}
	file.Decls = append(file.Decls, rootFn)

	// imports
	if len(ctx.ImportID) > 0 {
		specs := []ast.Spec{
			&ast.ImportSpec{
				Name: ast.NewIdent("dix"),
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: strconv.Quote("github.com/smtdfc/dix"),
				},
			},
		}
		for mod, alias := range ctx.ImportID {
			specs = append(specs, &ast.ImportSpec{
				Name: ast.NewIdent(alias),
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: strconv.Quote(mod),
				},
			})
		}
		importDecl := &ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: specs,
		}
		file.Decls = append([]ast.Decl{importDecl}, file.Decls...)
	}

	fset := token.NewFileSet()
	return ASTToString(fset, file)
}

// ASTToString converts an AST tree to its string representation.
func ASTToString(fset *token.FileSet, file *ast.File) (string, error) {
	var buf bytes.Buffer
	err := format.Node(&buf, fset, file)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// WriteGoFile converts and writes an AST tree into a Go source file.
func WriteGoFile(fset *token.FileSet, file *ast.File, outPath string) error {
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()
	return format.Node(out, fset, file)
}
