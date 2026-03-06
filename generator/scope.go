package generator

import (
	"fmt"
	"go/ast"
)

type Scope struct {
	Imports      map[string]*ast.Ident
	Names        map[string]*ast.Ident
	Counter      int
	UniqueIdents map[string]int
}

func (s *Scope) UniqueIdent(typeName string) *ast.Ident {
	if s.UniqueIdents == nil {
		s.UniqueIdents = make(map[string]int)
	}
	s.UniqueIdents[typeName]++
	name := fmt.Sprintf("%s%d", typeName, s.UniqueIdents[typeName]-1)
	return ast.NewIdent(name)
}

func (s *Scope) Import(pkg string) *ast.Ident {
	if ident, ok := s.Imports[pkg]; ok {
		return ident
	} else {
		s.Counter++
		// Dùng Sprintf để format chuẩn và KHÔNG có dấu xuống dòng
		s.Imports[pkg] = &ast.Ident{
			Name: fmt.Sprintf("pkg%d", s.Counter),
		}

		return s.Imports[pkg]
	}
}

func NewScope() *Scope {
	return &Scope{
		Counter: 0,
		Imports: make(map[string]*ast.Ident),
		Names:   make(map[string]*ast.Ident),
	}
}
