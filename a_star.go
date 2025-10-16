package main

import (
	"math"
)

const (
	GCross    = 10
	GDiagonal = 14
)

func heuristica(point *Punto, goal *Punto) float64 {
	return math.Abs(float64(point.x-goal.x)) + math.Abs(float64(point.y-goal.y))
}

func a_start(laberinto maze, start_point, goal *Punto) *Punto {
	var nodo_actual, nodo_busqueda, vecino *Punto
	var indice_menor_f int
	var bandera_vecino bool

	start_point.f = heuristica(start_point, goal)
	lista_abierta := []*Punto{start_point}

	f, c := laberinto.getShape()

	// creamos la lista cerra
	lista_cerrada := make([][]int, f)

	for i := range lista_abierta {
		lista_cerrada[i] = make([]int, c)
	}

	movs := []Punto{
		{x: 1, y: -1},
		{x: -1, y: -1},
		{x: -1, y: 0},
		{x: 1, y: 0},
		{x: 0, y: 1},
		{x: 1, y: 1},
		{x: -1, y: 1},
		{x: 0, y: -1},
	}

	for len(lista_abierta) > 0 {
		nodo_actual = lista_abierta[0] // sacamos el primer nodo

		indice_menor_f = 0
		for i := 1; i < len(lista_abierta); i++ {
			nodo_busqueda = lista_abierta[i]
			if nodo_busqueda.f < nodo_actual.f {
				nodo_actual = nodo_busqueda
				indice_menor_f = i
			}
		}
		// quitamos el nodo de la lista del nodo actual ( a este piunto es nodo con menor f)
		lista_abierta = append(lista_abierta[:indice_menor_f], lista_abierta[indice_menor_f+1:]...)
		lista_cerrada[nodo_actual.y][nodo_actual.x] = 1 // indicamos que ya hemos visitado el nodo

		if nodo_actual.x == goal.x && nodo_actual.y == goal.y {
			return nodo_actual
		}

		// calculamos los vecinos
		for _, mov := range movs {
			vecino = NewPoint(
				nodo_actual.x+mov.x,
				nodo_actual.y+mov.y,
				0,
				0,
			)

			// verificamos si el vecino calculado esta dentro del mapa
			if vecino.y > 0 &&
				vecino.x > 0 &&
				vecino.x < c &&
				vecino.y < f &&
				lista_cerrada[nodo_actual.y][nodo_actual.x] == 0 &&
				laberinto[nodo_actual.y][nodo_actual.x] == Transitable {
				// validamos si es un movimiento diagonal o vertical
				if (math.Abs(float64(vecino.x)) + math.Abs(float64(vecino.y))) == 2 {
					// es diagonal
					vecino.g = GDiagonal
				} else {
					vecino.g = GCross
				}

				vecino.f = vecino.g + heuristica(vecino, goal)

				// validamos  si es nodo esta en la lista o es conveniente agregarlo
				// es decir, ver si no ya existe un nodo como menor costo al vecino
				// calculado
				bandera_vecino = false
				for _, nodo := range lista_abierta {
					// nos preguntamos si ya existe el nodo , este tiene un costo menor
					if nodo.x == vecino.x && nodo.y == vecino.y && nodo.f <= vecino.f {
						bandera_vecino = true
						break
					}
				}
				// si no existe el nodo con f menor, lo agregamos
				if !bandera_vecino {
					lista_abierta = append(lista_abierta, vecino)
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
