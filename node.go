package main

import (
	"fmt"
	"slices"
)

// Node representa un cordenada en la mapa
// almacena valores necesarios para calcular movimientos
type Node struct {
	X, Y int
	// para calculo A*
	f      float64 // costo de transitar al nodo
	g      float64 // costo total mas heuristica
	parent *Node
}

func NewNode(x, y int) *Node {
	return &Node{
		X: x,
		Y: y,
	}
}

func (n *Node) BuildWay() []*Node {
	var camino []*Node

	nodoFinal := n
	for nodoFinal != nil {
		camino = append(camino, nodoFinal)
		nodoFinal = nodoFinal.parent
	}

	slices.Reverse(camino)

	return camino

}

func (n *Node) Copy(other *Node) {
	n.X = other.X
	n.Y = other.Y
}

func (n *Node) Clone() *Node {
	return &Node{
		X: n.X,
		Y: n.Y,
	}
}

func (n *Node) String() string {
	return fmt.Sprintf("(x: %v,Y: %v,f: %v)", n.X, n.Y, n.f)
}
