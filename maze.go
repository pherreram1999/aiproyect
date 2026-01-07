package main

import (
	"math/rand"
	"time"
)

type Maze [][]int

// Pos guarda una posición en la matriz del laberinto.
// Usamos Y (fila) Y X (columna) para mantener el mismo orden que en el código Python.
type Pos struct {
	Y int
	X int
}

// Genera un laberinto con loops internos
func NewMaze(ancho, alto int) Maze {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	m := newMazePerfect(ancho, alto, rng)
	braidDeadEnds(m, 0.35, rng)

	return m
}

// Laberinto base con un solo camino (perfect maze)
func newMazePerfect(ancho, alto int, rng *rand.Rand) Maze {
	if ancho%2 == 0 {
		ancho++
	}

	// Si el alto es par, lo incrementamos para que sea impar.
	if alto%2 == 0 {
		alto++
	}

	lab := make([][]int, alto)
	for y := range lab {
		lab[y] = make([]int, ancho)
		for x := range lab[y] {
			lab[y][x] = 1
		}
	}

	startX := rng.Intn(ancho/2)*2 + 1
	startY := rng.Intn(alto/2)*2 + 1
	lab[startY][startX] = 0

	var walls []Pos
	dirs := []Pos{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	for _, d := range dirs {
		walls = append(walls, Pos{startY + d.Y, startX + d.X})
	}

	for len(walls) > 0 {
		i := rng.Intn(len(walls))
		w := walls[i]
		walls = append(walls[:i], walls[i+1:]...)

		var c1, c2 Pos
		if w.Y%2 == 0 {
			c1 = Pos{w.Y - 1, w.X}
			c2 = Pos{w.Y + 1, w.X}
		} else {
			c1 = Pos{w.Y, w.X - 1}
			c2 = Pos{w.Y, w.X + 1}
		}

		if !inside(lab, c1) || !inside(lab, c2) {
			continue
		}

		if lab[c1.Y][c1.X] != lab[c2.Y][c2.X] {
			lab[w.Y][w.X] = 0

			var next Pos
			if lab[c1.Y][c1.X] == 1 {
				next = c1
			} else {
				next = c2
			}
			lab[next.Y][next.X] = 0

			// Añadimos los muros vecinos de la nueva celda a la lista de muros.
			// Recorremos las 4 direcciones para obtener los vecinos.
			for _, d := range dirs {
				ny, nx := next.Y+d.Y, next.X+d.X
				if insideXY(lab, ny, nx) && lab[ny][nx] == 1 {
					walls = append(walls, Pos{ny, nx})
				}
			}
		}
	}

	return lab
}

// Rompe callejones para crear loops internos
func braidDeadEnds(m Maze, prob float64, rng *rand.Rand) {
	h, w := len(m), len(m[0])
	dirs := []Pos{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	openNeighbors := func(y, x int) int {
		c := 0
		for _, d := range dirs {
			ny, nx := y+d.Y, x+d.X
			if insideXY(m, ny, nx) && m[ny][nx] == 0 {
				c++
			}
		}
		return c
	}

	for y := 1; y < h-1; y += 2 {
		for x := 1; x < w-1; x += 2 {
			if m[y][x] != 0 || openNeighbors(y, x) != 1 {
				continue
			}
			if rng.Float64() > prob {
				continue
			}

			jumps := []Pos{{-2, 0}, {2, 0}, {0, -2}, {0, 2}}
			for _, j := range jumps {
				ty, tx := y+j.Y, x+j.X
				wy, wx := (y+ty)/2, (x+tx)/2
				if insideXY(m, ty, tx) && m[ty][tx] == 0 && m[wy][wx] == 1 {
					m[wy][wx] = 0
					break
				}
			}
		}
	}
}

// Helpers
func inside(m Maze, p Pos) bool {
	return p.Y >= 0 && p.Y < len(m) && p.X >= 0 && p.X < len(m[0])
}

func insideXY(m Maze, y, x int) bool {
	return y >= 0 && y < len(m) && x >= 0 && x < len(m[0])
}

func (m Maze) GetShape() (int, int) {
	return len(m), len(m[0])
}

func (m Maze) Set(x, y, v int) {
	m[y][x] = v
}

func (m Maze) Get(x, y int) int {
	return m[y][x]
}
