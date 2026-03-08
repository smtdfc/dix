package generator

import (
	"github.com/smtdfc/dix/parser"
)

type Node struct {
	Provider *parser.Provider
	Deps     []*Node
}

type ProviderMap map[string]*parser.Provider

type Graph struct {
	Root *Node
}

func (g *Graph) Sort() ([]*parser.Provider, error) {
	var sorted []*parser.Provider
	status := make(map[*Node]int)

	var visit func(n *Node) error
	visit = func(n *Node) error {
		if status[n] == 1 {
			return NewGenerateError(
				ErrorDependencyResolve,
				"circular dependency detected",
				n.Provider.Name,
				"",
				nil,
			)
		}
		if status[n] == 2 {
			return nil
		}

		status[n] = 1

		for _, dep := range n.Deps {
			if err := visit(dep); err != nil {
				return err
			}
		}

		status[n] = 2
		sorted = append(sorted, n.Provider)
		return nil
	}

	if err := visit(g.Root); err != nil {
		return nil, err
	}

	return sorted, nil
}
func BuildGraph(root *parser.Provider, providerMap ProviderMap) (*Graph, error) {

	visited := make(map[string]*Node)

	var buildNode func(p *parser.Provider) (*Node, error)
	buildNode = func(p *parser.Provider) (*Node, error) {
		if p.Return == nil {
			return nil, NewGenerateError(
				ErrorValidation,
				"provider must declare exactly one return value",
				p.Name,
				"",
				nil,
			)
		}

		sig := p.Return.Type.Signature()

		if n, ok := visited[sig]; ok {
			return n, nil
		}

		if p.IsDisable {
			return nil, NewGenerateError(
				ErrorDependencyResolve,
				"disabled provider cannot be used as dependency or root",
				p.Name,
				"",
				nil,
			)
		}

		node := &Node{
			Provider: p,
			Deps:     make([]*Node, 0),
		}
		visited[sig] = node

		for _, dep := range p.Deps {
			depSig := dep.Type.Signature()

			childProvider, ok := providerMap[depSig]
			if !ok {
				return nil, NewGenerateError(
					ErrorDependencyResolve,
					"provider not found for dependency",
					p.Name,
					depSig,
					nil,
				)
			}

			childNode, err := buildNode(childProvider)
			if err != nil {
				return nil, NewGenerateError(
					ErrorGraphBuild,
					"failed to build provider graph",
					p.Name,
					depSig,
					err,
				)
			}
			node.Deps = append(node.Deps, childNode)
		}
		return node, nil
	}

	rootNode, err := buildNode(root)
	if err != nil {
		return nil, err
	}

	return &Graph{Root: rootNode}, nil
}
