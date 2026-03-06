package generator

import (
	"fmt"

	"github.com/smtdfc/dix/parser"
)

type Node struct {
	Composition *parser.Composition
	Deps        []*Node
}

type CompositionMap map[string]*parser.Composition

type Graph struct {
	Root *Node
}

func (g *Graph) Sort() []*parser.Composition {
	var sorted []*parser.Composition
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
		sorted = append(sorted, n.Composition)
	}

	visit(g.Root)
	return sorted
}

func BuildGraph(root *parser.Composition, compositionMap CompositionMap) (*Graph, error) {

	visited := make(map[string]*Node)

	var buildNode func(comp *parser.Composition) (*Node, error)
	buildNode = func(comp *parser.Composition) (*Node, error) {
		sig := comp.Return.Type.Signature()

		if n, ok := visited[sig]; ok {
			return n, nil
		}

		node := &Node{
			Composition: comp,
			Deps:        make([]*Node, 0),
		}
		visited[sig] = node

		for _, dep := range comp.Deps {
			depSig := dep.Type.Signature()

			childComp, ok := compositionMap[depSig]
			if !ok {
				return nil, fmt.Errorf("Provider not found: %s (Required by %s)", depSig, comp.Name)
			}

			childNode, err := buildNode(childComp)
			if err != nil {
				return nil, err
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
