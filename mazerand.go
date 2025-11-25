package main

import "math/rand/v2"

func Mazerand(mapa Maze) {
	counter := 0

	f, c := mapa.GetShape()

	for counter < NumAjolotes {
		// generamos un punto aleatorio
		x := rand.IntN(c)
		y := rand.IntN(f)
		if mapa.Get(x, y) == 0 {
			// es un camino transitable
			mapa.Set(x, y, 3) // el 3 indica que es un punto ajolote-point
			counter++
		}
	}
}
