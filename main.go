package main

import (
	"embed"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets
var assetsFS embed.FS

type Dimensiones struct {
	Alto     int
	Ancho    int
	Filas    int
	Columnas int
}

type MazeAssets struct {
	Wall  *ebiten.Image
	Floor *ebiten.Image
}

type Game struct {
	Maze        Maze
	Dimensiones *Dimensiones
	Player      *Player
	IsMoving    bool
	MazeAssets  *MazeAssets
	Enemy       *Enemy
}

// MovePlayer se encarga de crear de calcular las frames actuales Y las posiciones vectoriales
func (j *Game) MovePlayer() {
	p := j.Player

	// ✅ Primero, continuar movimiento en progreso
	if p.IsMoving {
		p.Moving()
	}

	// ✅ Luego, detectar nuevas teclas solo si NO está moviéndose
	if !p.IsMoving {
		if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
			p.MoveToUp()
		} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
			p.MoveToDown()
		} else if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
			p.MoveToLeft()
		} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
			p.MoveToRight()
		}
	}

	p.Tick() // Avanzar animaciones
}

func (j *Game) Update() error {
	j.MovePlayer()
	return nil
}

func (j *Game) DrawMaze(screen *ebiten.Image) {
	// dibujamos el laberitno (escenario)
	for f := 0; f < j.Dimensiones.Filas; f++ {
		for c := 0; c < j.Dimensiones.Columnas; c++ {
			y := float64(f * squareSize)
			x := float64(c * squareSize)

			var mazeAsset *ebiten.Image
			// Si el valor en el mapa es 1, es una pared
			if j.Maze[f][c] == 1 {
				// Pared - pintamos de gris oscuro
				mazeAsset = j.MazeAssets.Wall
			} else {
				// Camino - pintamos de gris claro
				mazeAsset = j.MazeAssets.Floor
			}

			imgOptions := &ebiten.DrawImageOptions{}
			// vector.FillRect(screen, x, y, squareSize, squareSize, colorCelda, false)

			imgOptions.GeoM.Translate(x, y)

			screen.DrawImage(mazeAsset, imgOptions)

		}
	}
}

func (j *Game) Draw(screen *ebiten.Image) {
	j.DrawMaze(screen)
	// dibujamos el jugaodor
	// lo colocamos en medio de la celda
	j.Player.DrawPlayer(screen)

}

func (j *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return j.Dimensiones.Ancho, j.Dimensiones.Alto
}

func main() {

	// cargamos los assets
	puntoInicial := NewNode(1, 1)

	middleX := float64(puntoInicial.X * squareSize)
	middleY := float64(puntoInicial.Y * squareSize)

	jugador := NewPlayer()

	jugador.MovingAnimation = NewAnimation(
		&AnimationOption{
			Assets:         assetsFS,
			Indexes:        [2]int{0, 9},
			TemplateString: "assets/nibbit_walking/f_000%d.png",
			Elapse:         3,
		},
	)

	jugador.StayAnimation = NewAnimation(
		&AnimationOption{
			Assets:         assetsFS,
			Indexes:        [2]int{0, 11},
			TemplateString: "assets/nibbit_staying/fs_%d.png",
			Elapse:         TPS * 0.2,
		},
	)

	startPosition := NewVector(middleX, middleY)

	jugador.CurrentPosition = startPosition.Clone()
	jugador.TargetPosition = startPosition.Clone() // clonamos para evitar escribir la misma direccion de memoria
	jugador.NodePosition = NewNode(1, 1)

	juego := &Game{
		Maze:        NewMaze(60, 35),
		Dimensiones: &Dimensiones{},
		Player:      jugador,
		MazeAssets: &MazeAssets{
			Floor: openAsset(assetsFS, "assets/floor.png"),
			Wall:  openAsset(assetsFS, "assets/wall.png"),
		},
	}

	// para que el jugador tenga acceso al los datos del juego
	juego.Player.Game = juego

	f, c := juego.Maze.GetShape()

	juego.Dimensiones.Alto = f * squareSize
	juego.Dimensiones.Ancho = c * squareSize
	juego.Dimensiones.Filas = f
	juego.Dimensiones.Columnas = c

	// cargamos al enemigo

	juego.Enemy = &Enemy{
		NodePosition: NewNode(c-1, f-1), // columnas, filas
	}

	juego.Enemy.Juego = juego

	ebiten.SetWindowSize(juego.Dimensiones.Ancho, juego.Dimensiones.Alto)
	ebiten.SetWindowTitle("Catch me!")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	if err := ebiten.RunGame(juego); err != nil {
		panic(err)
	}
}
