package main

import (
	"embed"
	"fmt"
	"image/color"
	"math"

	_ "image/gif"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	Transitable = 0

	squareSize = 15
	moveSpeed  = 2

	radioSize        = squareSize / 3
	squaredMoveSpeed = moveSpeed * squareSize

	TPS = 60
)

type Direccion int

const (
	DireccionArriba Direccion = iota
	DireccionDerecha
	DireccionAbajo
	DireccionIzquierda
)

func gradosARadianes(grados float64) float64 {
	return grados * math.Pi / 180
}

//go:embed assets
var assetsFS embed.FS

type Dimensiones struct {
	Alto     int
	Ancho    int
	Filas    int
	Columnas int
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

type PlayerAnimation struct {
	WalkingFrames     []*ebiten.Image
	CurrentFrameIndex int
	TickCounter       int
	TickElapse        int
	NumFrames         int
}

func (self *PlayerAnimation) GetCurrentFrame() *ebiten.Image {
	return self.WalkingFrames[self.CurrentFrameIndex]
}

type Player struct {
	position       *Vector
	targetPosition *Vector
	punto          *Punto
	Direccion      Direccion
	// assets
	Animation *PlayerAnimation
}

func (p *Player) obtenerDireccionAngulo() float64 {
	switch p.Direccion {
	case DireccionAbajo:
		return gradosARadianes(0) // 0°
	case DireccionIzquierda:
		return gradosARadianes(90) // 90°
	case DireccionArriba:
		return gradosARadianes(180) // 180°
	case DireccionDerecha:
		return gradosARadianes(270) // 270°
	}
	return gradosARadianes(0)

}

type Juego struct {
	Maze        maze
	Dimensiones *Dimensiones
	Player      *Player
	IsMoving    bool
}

func (j *Juego) MovePlayer() {
	// calculamos el frame del movimeinto basandonos en los ticks

	j.Player.Animation.TickCounter++ // aumentos el tick

	if j.Player.Animation.TickCounter > j.Player.Animation.TickElapse {
		j.Player.Animation.TickCounter = 0 // se renicia el contador de ticks
		j.Player.Animation.CurrentFrameIndex++
		if j.Player.Animation.CurrentFrameIndex >= j.Player.Animation.NumFrames {
			j.Player.Animation.CurrentFrameIndex = 0
		}
	}

	// calculamos el angulo segun su direccion

	if !j.IsMoving {
		puntoDestino := j.Player.punto.Clone()
		if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
			puntoDestino.y-- // sube un posicion arriba, el index anterior
			j.IsMoving = true
			j.Player.Direccion = DireccionArriba
		} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
			puntoDestino.y++ // baja un posicion, aumenta el indice
			j.IsMoving = true
			j.Player.Direccion = DireccionAbajo
		} else if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
			puntoDestino.x-- // nos vamos a la izquierda
			j.IsMoving = true
			j.Player.Direccion = DireccionIzquierda
		} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
			puntoDestino.x++ // nos vamos a la derecha
			j.IsMoving = true
			j.Player.Direccion = DireccionDerecha
		}
		// validamos si los nuevos movimientos estan dentro del mapa
		if (puntoDestino.y >= 0 && puntoDestino.y < j.Dimensiones.Filas) && (puntoDestino.x >= 0 && puntoDestino.x < j.Dimensiones.Columnas) {
			// estamos dentro del mapa, vemos si es un movmiento transitable
			if j.Maze[puntoDestino.y][puntoDestino.x] == 0 { // es transitable
				// colocamos el destino pero hacia el centro
				j.Player.targetPosition.x = float64(puntoDestino.x * squareSize)
				j.Player.targetPosition.y = float64(puntoDestino.y * squareSize)
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
	j.MovePlayer()
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
	// Obtenemos la posición actual del jugador
	// Estas coordenadas representan el CENTRO donde queremos dibujar
	jx := j.Player.position.x
	jy := j.Player.position.y
	//
	//// Creamos las opciones de transformación
	imgOptions := &ebiten.DrawImageOptions{}

	playerFrame := j.Player.Animation.GetCurrentFrame()
	bounds := playerFrame.Bounds()
	w := float64(bounds.Dx())
	h := float64(bounds.Dy())

	// Rotar desde el centro de la imagen
	// se tiene que centrar,para al momento de girar no salga de cuadro
	imgOptions.GeoM.Translate(-w/2, -h/2) // lo movemos hacia su origen desde el centro
	imgOptions.GeoM.Rotate(j.Player.obtenerDireccionAngulo())
	imgOptions.GeoM.Translate(w/2, h/2) // lo regresamos

	imgOptions.GeoM.Translate(
		jx, // Centramos horizontalmente
		jy, // Centramos verticalmente
	)
	//
	//// Dibujamos la imagen con todas las transformaciones aplicadas
	screen.DrawImage(playerFrame, imgOptions)
}

func (j *Juego) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return j.Dimensiones.Ancho, j.Dimensiones.Alto
}

func LoadWalkingAssets() []*ebiten.Image {
	var frames []*ebiten.Image
	// cargamos los assets de nibit caminanando
	for i := 0; i <= 9; i++ {
		f, err := assetsFS.Open(fmt.Sprintf("assets/nibbit_walking/f_000%d.png", i))
		if err != nil {
			panic(err)
		}

		frame, _, err := ebitenutil.NewImageFromReader(f)

		if err != nil {
			panic(err)
		}

		frames = append(frames, frame)
	}

	return frames
}

func main() {

	// cargamos los assets
	puntoInicial := &Punto{x: 1, y: 1}

	middleX := float64(puntoInicial.x * squareSize)
	middleY := float64(puntoInicial.y * squareSize)

	walkingAssets := LoadWalkingAssets()

	// player dimensions

	juego := &Juego{
		Maze:        CrearLaberintoPrim(60, 35),
		Dimensiones: &Dimensiones{},
		Player: &Player{
			punto:          &Punto{x: 1, y: 1},
			Direccion:      DireccionAbajo,
			position:       &Vector{middleX, middleY},
			targetPosition: &Vector{middleX, middleY},
			Animation: &PlayerAnimation{
				WalkingFrames: walkingAssets,
				TickElapse:    TPS * 0.2,
				NumFrames:     len(walkingAssets),
			},
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
