package main

import (
	"math"
)

const (
	GCross    = 10
	GDiagonal = 14
)

func heuristica(point *Node, goal *Node) float64 {
	return math.Abs(float64(point.X-goal.X)) + math.Abs(float64(point.Y-goal.Y))
}

func AStart(laberinto Maze, start_point, goal *Node) *Node {

	var nodoBusqueda *Node
	var indiceMenorF int
	var banderaVecino bool

	start_point.f = heuristica(start_point, goal)
	listaAbierta := []*Node{start_point}

	f, c := laberinto.GetShape()

	// creamos la lista cerra
	listaCerrada := make([][]int, f)

	for i := 0; i < f; i++ {
		listaCerrada[i] = make([]int, c)
	}

	moves := []Node{
		{X: 1, Y: -1},
		{X: -1, Y: -1},
		{X: -1, Y: 0},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},
		{X: -1, Y: 1},
		{X: 0, Y: -1},
	}

	for len(listaAbierta) > 0 {
		nodoActual := listaAbierta[0] // sacamos el primer nodo

		indiceMenorF = 0
		for i := 1; i < len(listaAbierta); i++ {
			nodoBusqueda = listaAbierta[i]
			if nodoBusqueda.f < nodoActual.f {
				nodoActual = nodoBusqueda
				indiceMenorF = i
			}
		}
		// quitamos el nodo de la lista del nodo actual ( a este piunto es nodo con menor f)
		listaAbierta = append(listaAbierta[:indiceMenorF], listaAbierta[indiceMenorF+1:]...)

		if nodoActual.X == goal.X && nodoActual.Y == goal.Y {
			return nodoActual
		}

		listaCerrada[nodoActual.Y][nodoActual.X] = 1 // indicamos que ya hemos visitado el nodo

		// calculamos los vecinos
		for _, mov := range moves {
			vecino := NewNode(
				nodoActual.X+mov.X,
				nodoActual.Y+mov.Y,
			)

			// verificamos si el vecino calculado esta dentro del mapa
			if vecino.Y >= 0 &&
				vecino.X >= 0 &&
				vecino.X < c &&
				vecino.Y < f &&
				listaCerrada[vecino.Y][vecino.X] == 0 &&
				// debemos considera los ajolote points como transitables
				(laberinto[vecino.Y][vecino.X] == Transitable || laberinto[vecino.Y][vecino.X] == AjolotePointType) {
				// validamos si es un movimiento diagonal o vertical
				// es el costo de traer desde nodo anterior, mas el nuevo costo de transitar
				if (math.Abs(float64(mov.X)) + math.Abs(float64(mov.Y))) == 2 {
					// es diagonal
					vecino.g = nodoActual.g + GDiagonal
				} else {
					vecino.g = nodoActual.g + GCross
				}

				vecino.f = vecino.g + heuristica(vecino, goal)
				vecino.parent = nodoActual

				// validamos  si es nodo esta en la lista o es conveniente agregarlo
				// es decir, ver si no ya existe un nodo como menor costo al vecino
				// calculado
				banderaVecino = false
				for _, nodo := range listaAbierta {
					// nos preguntamos si ya existe el nodo , este tiene un costo menor
					if nodo.X == vecino.X && nodo.Y == vecino.Y && nodo.f <= vecino.f {
						banderaVecino = true
						break
					}
				}
				// si no existe el nodo con f menor, lo agregamos
				if !banderaVecino {
					listaAbierta = append(listaAbierta, vecino)
				} else { // no nos intersa el vecino calculado lo limpiamos
					vecino = nil
				}
			} else {
				// de no ser el caso el nodo es descartado
				vecino = nil
			}
		}
	}
	return nil
}
