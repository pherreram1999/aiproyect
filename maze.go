package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Maze [][]int

// Pos guarda una posición en la matriz del laberinto.
// Usamos Y (fila) Y X (columna) para mantener el mismo orden que en el código Python.
type Pos struct {
	Y int // fila (coordenada vertical)
	X int // columna (coordenada horizontal)
}

// NewVector genera un laberinto usando una variante del algoritmo de Prim.
// Devuelve una matriz [][]int donde 1 = muro Y 0 = pasillo.
func NewMaze(ancho, alto int) Maze {
	// Si el ancho es par, lo incrementamos para que sea impar (bordes claros).
	if ancho%2 == 0 {
		ancho++
	}

	// Si el alto es par, lo incrementamos para que sea impar.
	if alto%2 == 0 {
		alto++
	}

	// Creamos la matriz del laberinto con 'alto' filas.
	laberinto := make([][]int, alto)
	// Para cada fila, creamos una slice de longitud 'ancho' inicializada con ceros por defecto.
	for y := range laberinto {
		laberinto[y] = make([]int, ancho)
		// Rellenamos explícitamente con 1 para representar muros.
		for x := range laberinto[y] {
			laberinto[y][x] = 1 // 1 representa muro
		}
	}

	// Semilla aleatoria basada en el tiempo actual para resultados distintos en cada ejecución.
	rand.Seed(time.Now().UnixNano())

	// Elegimos un punto inicial aleatorio con coordenadas impares:
	// rand.Intn(ancho/2) devuelve un valor en [0, ancho/2), lo multiplicamos por 2 Y sumamos 1 => {1,3,5,...}
	inicioX := rand.Intn(ancho/2)*2 + 1
	inicioY := rand.Intn(alto/2)*2 + 1

	// Marcamos la celda inicial como pasillo (0).
	laberinto[inicioY][inicioX] = 0

	// Definimos el tipo lista de muros (frontera). Usamos Pos para almacenar coordenadas (Y,X).
	var muros []Pos

	// Definimos las 4 direcciones como desplazamientos en Y,X.
	dirs := []Pos{
		{Y: -1, X: 0}, // arriba
		{Y: 1, X: 0},  // abajo
		{Y: 0, X: -1}, // izquierda
		{Y: 0, X: 1},  // derecha
	}

	// Añadimos los muros vecinos (adyacentes) de la celda inicial a la lista de muros.
	// Recorremos cada dirección Y calculamos la posición vecina.
	for _, d := range dirs {
		y := inicioY + d.Y // fila vecina
		x := inicioX + d.X // columna vecina
		// Comprobamos que la posición esté dentro de los límites.
		if y >= 0 && y < alto && x >= 0 && x < ancho {
			// Añadimos la posición a la lista de muros.
			muros = append(muros, Pos{Y: y, X: x})
		}
	}

	// Bucle principal: procesamos mientras haya muros en la lista.
	for len(muros) > 0 {
		// Elegimos un índice aleatorio entre 0 Y len(muros)-1.
		i := rand.Intn(len(muros))
		// Obtenemos el muro seleccionado.
		muro := muros[i]
		muroY, muroX := muro.Y, muro.X

		// Eliminamos el muro seleccionado de la slice (mantenemos el orden no importantísimo).
		// Lo hacemos concatenando la parte izquierda Y la parte derecha sin el elemento i.
		muros = append(muros[:i], muros[i+1:]...)

		// Determinamos las dos celdas opuestas separadas por este muro.
		// Si el muro está en una fila par => muro vertical (entre celdas verticales).
		// Si el muro está en una fila impar => muro horizontal (entre celdas horizontales).
		var celda1 Pos
		var celda2 Pos

		// Si la coordenada Y del muro es par, el muro es horizontal en la malla de celdas (separa arriba/abajo).
		// (esto sigue la convención de usar celdas en posiciones impares, muros en posiciones pares).
		if muroY%2 == 0 {
			celda1 = Pos{Y: muroY - 1, X: muroX} // celda arriba del muro
			celda2 = Pos{Y: muroY + 1, X: muroX} // celda abajo del muro
		} else {
			// Si Y es impar, entonces el muro separa izquierda/derecha.
			celda1 = Pos{Y: muroY, X: muroX - 1} // celda izquierda del muro
			celda2 = Pos{Y: muroY, X: muroX + 1} // celda derecha del muro
		}

		// Verificamos límites: si alguna de las celdas está fuera, el muro toca el borde exterior Y lo ignoramos.
		if celda1.Y < 0 || celda1.X < 0 || celda2.Y < 0 || celda2.X < 0 ||
			celda1.Y >= alto || celda1.X >= ancho || celda2.Y >= alto || celda2.X >= ancho {
			continue // pasa a la siguiente iteración
		}

		// Comprobamos si cada celda es muro (no visitada) o ya es pasillo (visitada).
		celda1EsMuro := laberinto[celda1.Y][celda1.X] == 1
		celda2EsMuro := laberinto[celda2.Y][celda2.X] == 1

		// Si exactamente una de las celdas es muro (es decir, c1 != c2),
		// entonces podemos abrir el muro para conectar la celda visitada con la no visitada.
		if celda1EsMuro != celda2EsMuro {
			// Convertimos el muro en pasillo (abrimos el paso).
			laberinto[muroY][muroX] = 0

			// Determinamos cuál de las dos celdas será la nueva celda abierta (la que era muro).
			var nueva Pos
			if celda1EsMuro {
				nueva = celda1 // abrimos celda1 si era muro
			} else {
				nueva = celda2 // abrimos celda2 si era muro
			}

			// Marcamos la nueva celda como pasillo.
			laberinto[nueva.Y][nueva.X] = 0

			// Añadimos los muros vecinos de la nueva celda a la lista de muros.
			// Recorremos las 4 direcciones para obtener los vecinos.
			for _, d := range dirs {
				ny := nueva.Y + d.Y // fila del vecino
				nx := nueva.X + d.X // columna del vecino

				// Verificamos que el vecino esté dentro de los límites Y que sea un muro.
				if ny >= 0 && ny < alto && nx >= 0 && nx < ancho && laberinto[ny][nx] == 1 {
					// Evitamos añadir duplicados: comprobamos si ya está en la lista 'muros'.
					duplicado := false
					for _, m := range muros {
						if m.Y == ny && m.X == nx {
							duplicado = true
							break
						}
					}
					// Si no está duplicado, lo añadimos.
					if !duplicado {
						muros = append(muros, Pos{Y: ny, X: nx})
					}
				}
			}
		}
		// Si ambas celdas ya son pasillo o ambas son muro, no hacemos nada (no conectamos).
	}

	// Devolvemos la matriz completa con muros Y pasillos.
	return laberinto
}

// ImprimirLaberinto muestra por consola el laberinto usando emojis: 🧱 = muro, espacio = pasillo.
func ImprimirLaberinto(laberinto [][]int) {
	// Recorremos cada fila.
	for _, fila := range laberinto {
		// Recorremos cada celda de la fila.
		for _, celda := range fila {
			// Si la celda vale 1 imprimimos un emoji de muro.
			if celda == 1 {
				fmt.Print("🧱")
			} else {
				// Si es 0 imprimimos dos espacios para dar aspecto de pasillo.
				fmt.Print("  ")
			}
		}
		// Saltamos de línea después de cada fila.
		fmt.Println()
	}
}

func (m Maze) GetShape() (int, int) {
	filas := len(m)
	cols := 0
	if filas > 0 {
		cols = len(m[0])
	}
	return filas, cols
}
