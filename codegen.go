package dix

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"strconv"
)

func BuildOrder(container map[string]*Factory) ([]string, error) {
	indegree := map[string]int{}
	graph := map[string][]string{}

	// Build indegree + adjacency
	for alias, f := range container {
		if _, ok := indegree[alias]; !ok {
			indegree[alias] = 0
		}
		for _, dep := range f.Deps {
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

	return order, nil
}

type GenContext struct {
	ImportID  map[string]string // module path -> alias
	Container map[string]string // alias -> var id
	Counter   int
}

func (c *GenContext) GenUID() string {
	c.Counter++
	return "id_" + strconv.Itoa(c.Counter)
}

func (c *GenContext) ResolveImport(module string) string {
	if _, ok := c.ImportID[module]; !ok {
		c.ImportID[module] = c.GenUID()
	}
	return c.ImportID[module]
}

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

func GenerateCode(pkg string, config *DIConfig) (string, error) {
	ctx := &GenContext{
		ImportID:  make(map[string]string),
		Container: make(map[string]string),
		Counter:   0,
	}

	order, err := BuildOrder(config.Container)
	if err != nil {
		return "", err
	}

	file := &ast.File{
		Name:  ast.NewIdent("dix"),
		Decls: []ast.Decl{},
	}

	stmts := []ast.Stmt{}
	var depIdents []ast.Expr

	for _, item := range order {
		factory := config.Container[item]
		id := ctx.GenUID()
		ctx.Container[item] = id

		args := []ast.Expr{}
		for _, dep := range factory.Deps {
			args = append(args, generateDepExpr(ctx, dep, config))
		}

		modAlias := ctx.ResolveImport(factory.Module)

		// id_x := modAlias.Func(args...)
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

		stmts = append(stmts, assign)
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

	// import
	if len(ctx.ImportID) > 0 {
		specs := []ast.Spec{
			&ast.ImportSpec{
				Name: ast.NewIdent("dix"),
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "github.com/smtdfc/dix",
				},
			},
		}
		for mod, alias := range ctx.ImportID {
			specs = append(specs, &ast.ImportSpec{
				Name: ast.NewIdent(alias),
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: `"` + mod + `"`,
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

func ASTToString(fset *token.FileSet, file *ast.File) (string, error) {
	var buf bytes.Buffer
	err := format.Node(&buf, fset, file)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func WriteGoFile(fset *token.FileSet, file *ast.File, outPath string) error {
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()
	return format.Node(out, fset, file)
}
