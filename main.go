package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	squareSize = 20
)

type Dimensiones struct {
	Alto     int
	Ancho    int
	Filas    int
	Columnas int
}

type Player struct {
	X, Y int
}

type Juego struct {
	Maze        maze
	Dimensiones *Dimensiones
	Player      *Player
}

func (j *Juego) Update() error {
	// detectactos las teclas
	jx, jy := j.Player.X, j.Player.Y
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		jy-- // sube un posicion arriba, el index anterior
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		jy++ // baja un posicion arriba, aumenta el indice
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		jx-- // nos vamos a la izquierda
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		jx++ // nos vamos a la derecha
	}
	// validamos si los nuevos movimientos estan dentro del mapa

	if (jy > 0 && jy < j.Dimensiones.Filas) && (jx > 0 && jx < j.Dimensiones.Columnas) {
		// estamos dentro del mapa, vemos si es un movmiento transitable
		if j.Maze[jy][jx] == 0 { // es transitable
			j.Player.X = jx
			j.Player.Y = jy
		}
	}
	return nil
}

func (j *Juego) Draw(screen *ebiten.Image) {
	// dibujamos el laberitno (escenario)
	for f := 0; f < j.Dimensiones.Filas; f++ {
		for c := 0; c < j.Dimensiones.Columnas; c++ {
			y := float32(f * squareSize)
			x := float32(c * squareSize)

			var colorCelda color.Color
			// Si el valor en el mapa es 1, es una pared
			if j.Maze[f][c] == 1 {
				// Pared - pintamos de gris oscuro
				colorCelda = color.RGBA{60, 60, 60, 255}
			} else {
				// Camino - pintamos de gris claro
				colorCelda = color.RGBA{200, 200, 200, 255}
			}

			vector.FillRect(screen, x, y, squareSize, squareSize, colorCelda, false)

		}
	}
	// dibujamos el jugaodor

	radio := float32(squareSize) / 3 // tamaño del circulo
	// lo colocamos en medio de la celda
	jx := float32(j.Player.X*squareSize) + (squareSize / 2)
	jy := float32(j.Player.Y*squareSize) + (squareSize / 2)

	vector.FillCircle(screen, jx, jy, radio, color.RGBA{R: 50, G: 200, B: 50, A: 255}, true)
}

func (j *Juego) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return j.Dimensiones.Ancho, j.Dimensiones.Alto
}

func main() {
	juego := &Juego{
		Maze:        getMatriz(),
		Dimensiones: &Dimensiones{},
		Player:      &Player{1, 1},
	}

	f, c := juego.Maze.getShape()

	juego.Dimensiones.Alto = f * squareSize
	juego.Dimensiones.Ancho = c * squareSize
	juego.Dimensiones.Filas = f
	juego.Dimensiones.Columnas = c

	ebiten.SetWindowSize(juego.Dimensiones.Ancho, juego.Dimensiones.Alto)
	ebiten.SetWindowTitle("Catch me!")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	if err := ebiten.RunGame(juego); err != nil {
		panic(err)
	}
}
