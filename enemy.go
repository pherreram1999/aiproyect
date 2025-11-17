package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

type Enemy struct {
	Animation      *Animation
	NodePosition   *Node
	VectorPosition *Vector
	Elapse         int
	TickCounter    int
	Juego          *Game
	PathIndex      int
	Path           []*Node
}

// CalculatePath actualiza el path hacia el jugador
func (e *Enemy) CalculatePath() {
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

	imgOptions.GeoM.Translate(e.VectorPosition.X, e.VectorPosition.Y)
	frame := e.Animation.GetFrame()

	screen.DrawImage(frame, imgOptions)
}

func (e *Enemy) GetCurrentPathNode() *Node {
	return e.Path[e.PathIndex]
}

func (e *Enemy) UpdateVectorPosition() {
	e.VectorPosition = NewVector(
		float64(e.NodePosition.X*squareSize),
		float64(e.NodePosition.Y*squareSize),
	)
}

func (e *Enemy) Tick() {
	e.Animation.Tick()
	e.TickCounter++ // este tick es para avanzar el calculo de la ia

	if e.TickCounter > e.Elapse {
		// si se pasa, avanzamos un cuadrando al camino
		if e.PathIndex < len(e.Path)-1 {
			e.PathIndex++ // avanzamos un lugar en la ruta
		} else {
			// ha este punto se llega final de la ruta,
			// ahora el punto final
			e.NodePosition = e.Path[e.PathIndex]
			e.PathIndex = 0 // reniciamos el contador del path
			e.CalculatePath()
		}
		e.TickCounter = 0

		// actualizamos su posicion actual
		e.NodePosition = e.GetCurrentPathNode()
		// recalculamos los vectores
		e.UpdateVectorPosition() // en base a la nueva posicion actualizada
	}

}
