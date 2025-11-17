package main

import (
	"fmt"
	"log"
)

type Enemy struct {
	Animation    *Animation
	NodePosition *Node
	Elapse       int
	Tick         int
	Juego        *Game
	Ruta         []*Node
}

func (e *Enemy) CalcularRuta() {
	maze := e.Juego.Maze
	meta := e.Juego.Player.NodePosition
	nodoMeta := AStart(maze, e.NodePosition, meta)

	if nodoMeta == nil {
		log.Fatalln("Meta no calulada")
	}
	e.Ruta = nodoMeta.BuildWay()

	fmt.Println(e.Ruta)
}
