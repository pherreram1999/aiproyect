package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

type Enemy struct {
	Animation             *Animation // animacion de los sprites
	NodePosition          *Node      // indica la posicion dentro del mapa
	VectorCurrentPosition *Vector2d  // los vectores son utilizados para calcular un desplazamiento suave
	VectorTargetPosition  *Vector2d
	Elapse                int // lapso de tiempo en que se realiza el calculo del posicion del jugador
	ElapseDecrement       int // decrementos en avance en la reducion de lapso
	TickCounter           int
	Juego                 *Game
	PathIndex             int
	Path                  []*Node
	IsMoving              bool
}

type Enemys []*Enemy

// CalculatePath actualiza el path hacia el jugador
func (e *Enemy) CalculatePath() {
	// tenemos que liberar todos las referencias de memoria en camino previo
	for _, c := range e.Path {
		c.parent = nil
	}

	maze := e.Juego.Maze
	meta := e.Juego.Player.NodePosition
	nodoMeta := AStart(maze, e.NodePosition, meta)

	if nodoMeta == nil {
		log.Fatalln("Meta no calulada")
	}
	e.Path = nodoMeta.BuildWay()
}

func (e *Enemy) Draw(screen *ebiten.Image) {
	imgOptions := &ebiten.DrawImageOptions{}
	imgOptions.GeoM.Translate(e.VectorCurrentPosition.X, e.VectorCurrentPosition.Y)
	frame := e.Animation.GetFrame()

	screen.DrawImage(frame, imgOptions)
}

func (e *Enemy) GetCurrentPathNode() *Node {
	return e.Path[e.PathIndex]
}

func (e *Enemy) UpdateVectorTargetPosition() {
	e.VectorTargetPosition = NewVector(
		float64(e.NodePosition.X*squareSize),
		float64(e.NodePosition.Y*squareSize),
	)
}

// Tick determina los avances en lapsos de avance
func (e *Enemy) Tick() {
	e.Animation.Tick()

	if !e.IsMoving {
		// si no esta movmiento, calcualmos el siguiente paso
		e.TickCounter++ // este tick es para avanzar el calculo de la ia
		if e.TickCounter > e.Elapse {
			// si se pasa, avanzamos un cuadrando al camino
			e.TickCounter = 0
			e.IsMoving = true
			// calculamos a cada paso la ruta al enemigo
			e.CalculatePath()
			// dado el path se actualiza cada elapse,
			// siempre se avanza al segundo elemento de la ruta
			e.NodePosition = e.Path[e.PathIndex]

			e.UpdateVectorTargetPosition()
		}
	}

	if e.IsMoving { // si se esta movimiendo
		// calculamos su desplazamiento
		// vemos su distancia
		dir := e.VectorTargetPosition.Sub(e.VectorCurrentPosition)

		dist := dir.SquaredDistance()

		// la distancia es menos que la velocidad de desplazamiento, lo colocamos con su destino
		if dist <= squaredMoveSpeed {
			e.VectorCurrentPosition.X = e.VectorTargetPosition.X
			e.VectorCurrentPosition.Y = e.VectorTargetPosition.Y
			e.IsMoving = false
			return // ya no se mueve, no calculamos deplazamiento
		}

		uni := dir.Normalize()

		plus := uni.MultiplyByScalar(moveSpeed) // aumentar magnitud

		e.VectorCurrentPosition.X += plus.X
		e.VectorCurrentPosition.Y += plus.Y

	}

}
