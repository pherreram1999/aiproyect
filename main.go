package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	squareSize = 20
	moveSpeed  = 2

	radioSize        = squareSize / 3
	squaredMoveSpeed = moveSpeed * squareSize
)

type Dimensiones struct {
	Alto     int
	Ancho    int
	Filas    int
	Columnas int
}

type Punto struct {
	x, y int
}

func (self *Punto) Copy(other *Punto) {
	self.x = other.x
	self.y = other.y
}

func (self *Punto) Clone() *Punto {
	return &Punto{
		x: self.x,
		y: self.y,
	}
}

type Vector struct {
	x, y float64
}

func (self *Vector) Add(other *Vector) *Vector {
	return &Vector{
		x: self.x + other.x,
		y: self.y + other.y,
	}
}

func (self *Vector) Sub(other *Vector) *Vector {
	return &Vector{
		x: self.x - other.x,
		y: self.y - other.y,
	}
}

func (self *Vector) SquaredDistance() float64 {
	return self.x*self.x + self.y*self.y
}

func (self *Vector) Distance() float64 {
	return math.Sqrt(self.SquaredDistance())
}

func (self *Vector) Normalize() *Vector {
	mod := self.Distance()
	return &Vector{
		x: self.x / mod,
		y: self.y / mod,
	}
}

func (self *Vector) MultiplyByScalar(scalar float64) *Vector {
	return &Vector{
		x: self.x * scalar,
		y: self.y * scalar,
	}
}

func (self *Vector) String() string {
	return fmt.Sprintf("(%f.2,%f.2)", self.x, self.y)
}

type Player struct {
	position       *Vector
	targetPosition *Vector
	punto          *Punto
}

type Juego struct {
	Maze        maze
	Dimensiones *Dimensiones
	Player      *Player
	IsMoving    bool
}

func (j *Juego) Move() {
	if !j.IsMoving {
		puntoDestino := j.Player.punto.Clone()
		if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
			puntoDestino.y-- // sube un posicion arriba, el index anterior
			j.IsMoving = true
		} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
			puntoDestino.y++ // baja un posicion arriba, aumenta el indice
			j.IsMoving = true
		} else if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
			puntoDestino.x-- // nos vamos a la izquierda
			j.IsMoving = true
		} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
			puntoDestino.x++ // nos vamos a la derecha
			j.IsMoving = true
		}
		// validamos si los nuevos movimientos estan dentro del mapa

		if (puntoDestino.y >= 0 && puntoDestino.y < j.Dimensiones.Filas) && (puntoDestino.x >= 0 && puntoDestino.x < j.Dimensiones.Columnas) {
			// estamos dentro del mapa, vemos si es un movmiento transitable
			if j.Maze[puntoDestino.y][puntoDestino.x] == 0 { // es transitable
				j.Player.targetPosition.x = float64((puntoDestino.x * squareSize) + (squareSize / 2))
				j.Player.targetPosition.y = float64((puntoDestino.y * squareSize) + (squareSize / 2))
				j.Player.punto.Copy(puntoDestino)
				puntoDestino = nil // para garbage colector
			}

		}
	}

	// en cada frame hacemos el calculo del movimiento del vector
	if j.IsMoving {
		// obtenemos la direccion
		dir := j.Player.targetPosition.Sub(j.Player.position)
		// obtenemos distancia
		dist := dir.SquaredDistance()
		// comparamos mientras no segumos moviento hacia el objetivo
		// las ditancia entre ellos se hace menor
		// hasta el punto_destino que es menor que el los pixeles de movimiento
		// en ese punto_destino no detenemos
		if dist > squaredMoveSpeed {
			// normalizamos para solo aumentar en pasos el avance
			uni := dir.Normalize()
			// multiplicamos por le velocidad para conseguir las nuevas coordenadas
			target := uni.MultiplyByScalar(moveSpeed)
			j.Player.position = j.Player.position.Add(target)
			// para el garbarge colector
			target = nil
			uni = nil
			dir = nil
		} else {
			j.Player.position.x = j.Player.targetPosition.x
			j.Player.position.y = j.Player.targetPosition.y
			j.IsMoving = false
		}

	}
}

func (j *Juego) Update() error {
	j.Move()
	return nil
}

func (j *Juego) DrawMaze(screen *ebiten.Image) {
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
}

func (j *Juego) Draw(screen *ebiten.Image) {
	j.DrawMaze(screen)
	// dibujamos el jugaodor
	// lo colocamos en medio de la celda
	jx := float32(j.Player.position.x)
	jy := float32(j.Player.position.y)

	vector.FillCircle(screen, jx, jy, radioSize, color.RGBA{R: 50, G: 200, B: 50, A: 255}, true)
}

func (j *Juego) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return j.Dimensiones.Ancho, j.Dimensiones.Alto
}

func main() {
	puntoInicial := &Punto{1, 1}

	middleX := float64(puntoInicial.x*squareSize + (squareSize / 2))
	middleY := float64(puntoInicial.y*squareSize + (squareSize / 2))

	juego := &Juego{
		Maze:        CrearLaberintoPrim(60, 35),
		Dimensiones: &Dimensiones{},
		Player: &Player{
			punto:          &Punto{1, 1},
			position:       &Vector{middleX, middleY},
			targetPosition: &Vector{middleX, middleY},
		},
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
