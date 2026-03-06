package generator

import (
	"go/ast"
	"go/token"

	"github.com/smtdfc/dix/parser"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(metadata *parser.Metadata) (string, error) {
	_ = token.NewFileSet()

	file := &ast.File{
		Name: ast.NewIdent("generated"),
	}

	stmt := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("res")},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: ast.NewIdent("NewService"),
				Args: []ast.Expr{
					ast.NewIdent("db"),
				},
			},
		},
	}

	fn := &ast.FuncDecl{
		Name: ast.NewIdent("Init"),
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{stmt},
		},
	}

	file.Decls = append(file.Decls, fn)

	return "", nil
}
