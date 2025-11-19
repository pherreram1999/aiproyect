package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Direction int

const (
	DirectionUp Direction = iota
	DirectionRight
	DirectionDown
	DirectionLeft
)

type Player struct {
	StayAnimation    *Animation
	MovingAnimation  *Animation
	IsMoving         bool
	CurrentPosition  *Vector
	TargetPosition   *Vector
	NodePosition     *Node
	CurrentDirection Direction
	Points           uint
	// valores para validar movimientos en el mapa
	Mapa Maze
	// apuntamos al padre
	Game *Game
}

func NewPlayer() *Player {
	return &Player{
		CurrentDirection: DirectionDown,
	}
}

// movimientos del jugador

// hacia arriba

func (player *Player) validNode(node *Node) bool {
	mapa := player.Game.Maze
	f, c := player.Game.Dimensiones.Filas, player.Game.Dimensiones.Columnas
	return (node.Y > 0 && node.Y < f) &&
		(node.X > 0 && node.X < c) &&
		// considerar los ajolotes pesos
		(mapa[node.Y][node.X] == Transitable || mapa[node.Y][node.X] == AjolotePointType)
}

func (player *Player) Moving() {
	if !player.IsMoving {
		return
	}

	// Obtenemos la dirección
	dir := player.TargetPosition.Sub(player.CurrentPosition)
	// Obtenemos distancia al cuadrado
	dist := dir.SquaredDistance()

	// Si hemos llegado al destino
	// implica que la distancia entre ellos infima
	if dist <= squaredMoveSpeed {
		player.CurrentPosition.X = player.TargetPosition.X
		player.CurrentPosition.Y = player.TargetPosition.Y
		player.IsMoving = false
		return
	}

	// Normalizamos para obtener la dirección unitaria
	uni := dir.Normalize()
	// Multiplicamos por la velocidad
	movePlus := uni.MultiplyByScalar(moveSpeed)

	// ACTUALIZAR directamente las coordenadas, no crear nuevo vector
	player.CurrentPosition.X += movePlus.X
	player.CurrentPosition.Y += movePlus.Y
}

func (player *Player) Move(yMove, xMove int, direction Direction) {
	// Si ya está moviéndose, no hacer nada
	if player.IsMoving {
		return
	}

	// Clonamos el nodo para validar
	targetNode := player.NodePosition.Clone()
	targetNode.Y += yMove
	targetNode.X += xMove

	// Solo cambiar IsMoving si el movimiento es válido
	if player.validNode(targetNode) {
		player.IsMoving = true
		player.CurrentDirection = direction

		// Actualizar target position
		player.TargetPosition.X = float64(targetNode.X * squareSize)
		player.TargetPosition.Y = float64(targetNode.Y * squareSize)
		player.NodePosition = targetNode
	}
}

func (player *Player) MoveToUp() {
	player.Move(-1, 0, DirectionUp)
}

func (player *Player) MoveToDown() {
	player.Move(1, 0, DirectionDown)
}

func (player *Player) MoveToLeft() {
	player.Move(0, -1, DirectionLeft)
}

func (player *Player) MoveToRight() {
	player.Move(0, 1, DirectionRight)
}

func (player *Player) Tick() {
	if player.IsMoving {
		player.MovingAnimation.Tick()
	} else {
		player.StayAnimation.Tick()
	}
}

func (player *Player) GetSpriteFrame() *ebiten.Image {
	if player.IsMoving {
		return player.MovingAnimation.GetFrame()
	}
	return player.StayAnimation.GetFrame()
}

func gradosARadianes(grados float64) float64 {
	return grados * math.Pi / 180
}

func (p *Player) GetArg() float64 {
	switch p.CurrentDirection {
	case DirectionDown:
		return gradosARadianes(0) // 0°
	case DirectionLeft:
		return gradosARadianes(90) // 90°
	case DirectionUp:
		return gradosARadianes(180) // 180°
	case DirectionRight:
		return gradosARadianes(270) // 270°
	}
	return gradosARadianes(0)

}

func (player *Player) DrawPlayer(screen *ebiten.Image) {
	// Obtenemos la posición actual del jugador
	// Estas coordenadas representan el CENTRO donde queremos dibujar
	jx := player.CurrentPosition.X
	jy := player.CurrentPosition.Y
	//
	//// Creamos las opciones de transformación
	imgOptions := &ebiten.DrawImageOptions{}

	playerFrame := player.GetSpriteFrame()
	bounds := playerFrame.Bounds()
	w := float64(bounds.Dx())
	h := float64(bounds.Dy())

	// Rotar desde el centro de la imagen
	// se tiene que centrar,para al momento de girar no salga de cuadro
	imgOptions.GeoM.Translate(-w/2, -h/2) // lo movemos hacia su origen desde el centro
	imgOptions.GeoM.Rotate(player.GetArg())
	imgOptions.GeoM.Translate(w/2, h/2) // lo regresamos

	imgOptions.GeoM.Translate(
		jx, // Centramos horizontalmente
		jy, // Centramos verticalmente
	)
	//
	//// Dibujamos la imagen con todas las transformaciones aplicadas
	screen.DrawImage(playerFrame, imgOptions)
}
