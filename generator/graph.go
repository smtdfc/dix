package generator

import "github.com/smtdfc/dix/parser"

type Node struct {
	Provider *parser.Provider
	Deps     []*Node
}

type ProviderMap map[string]*parser.Provider

type Graph struct {
	Root *Node
}

func (g *Graph) Sort() []*parser.Provider {
	var sorted []*parser.Provider
	visited := make(map[*Node]bool)

	var visit func(n *Node)
	visit = func(n *Node) {
		if visited[n] {
			return
		}

		for _, dep := range n.Deps {
			visit(dep)
		}

		visited[n] = true
		sorted = append(sorted, n.Provider)
	}

	visit(g.Root)
	return sorted
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
