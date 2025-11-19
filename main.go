package main

import (
	"embed"
	"fmt"
	"image/color"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type State int

const (
	PlayingState State = iota
	GameOverState
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
	Wall             *ebiten.Image
	Floor            *ebiten.Image
	AjoloteAnimation *Animation
}

type Font struct {
	Face    *text.GoTextFace
	Options *text.DrawOptions
}
type Game struct {
	Maze        Maze
	Dimensiones *Dimensiones
	Player      *Player
	IsMoving    bool
	MazeAssets  *MazeAssets
	Enemy       *Enemy
	State       State
	Font        *Font
}

// MovePlayer se encarga de crear de calcular las frames actuales Y las posiciones vectoriales
func (j *Game) MovePlayer() {
	p := j.Player

	// continuar movimiento en progreso
	if p.IsMoving {
		p.Moving()
	}

	// detectar nuevas teclas solo si NO está moviéndose
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

	// tenemos que validar si el nodo si encuentra es un ajolote punto

	if j.Maze.Get(j.Player.NodePosition.X, j.Player.NodePosition.Y) == AjolotePointType {
		p.Points += AjolotePointValue

		j.Maze.Set(j.Player.NodePosition.X, j.Player.NodePosition.Y, Transitable) // indicamos que ya solo es camino
	}
}

func (j *Game) MoveEnemy() {
	e := j.Enemy
	e.Tick() // avanzar animaciones del enemigo
}

func GameOver() {

}

func (j *Game) DrawMaze(screen *ebiten.Image) {
	for f := 0; f < j.Dimensiones.Filas; f++ {
		for c := 0; c < j.Dimensiones.Columnas; c++ {
			y := float64(f * squareSize)
			x := float64(c * squareSize)

			var mazeAsset *ebiten.Image
			// Si el valor en el mapa es 1, es una pared
			celda := j.Maze[f][c]
			if celda == 1 {
				mazeAsset = j.MazeAssets.Wall
			} else if celda == 0 {
				mazeAsset = j.MazeAssets.Floor
			} else if celda == AjolotePointType {
				mazeAsset = j.MazeAssets.AjoloteAnimation.GetFrame()
			}

			imgOptions := &ebiten.DrawImageOptions{}
			// vector.FillRect(screen, x, y, squareSize, squareSize, colorCelda, false)

			imgOptions.GeoM.Translate(x, y)

			// en caso de que sea un ajolote point, lo tenemos que mezclar un tectura de camino

			if celda == AjolotePointType {
				screen.DrawImage(j.MazeAssets.Floor, imgOptions)
			}

			screen.DrawImage(mazeAsset, imgOptions)
		}
	}
}

func (j *Game) Update() error {
	if j.State == PlayingState {
		j.MovePlayer()
		j.MoveEnemy()
		// validamos si tanto el enemigo como el jugador llegaron a colisionar si estan en
		// en el mismo punto (nodo)

		if j.Enemy.NodePosition.Equal(j.Player.NodePosition) {
			// indicamos que tenemos que acabar el juego

		}

		// validamos si ya tiene todos lo puntos recoltados

		if j.Player.Points == MaxAjolotePoints {
		}
	}

	return nil
}

func (j *Game) Draw(screen *ebiten.Image) {

	if j.State == PlayingState {
		j.DrawMaze(screen)
		// dibujamos el jugaodor
		// lo colocamos en medio de la celda
		j.Player.DrawPlayer(screen)

		j.Enemy.Draw(screen)

		// animacion para los ajolote poins
		j.MazeAssets.AjoloteAnimation.Tick()

		// dibujamos le puntaje
		text.Draw(screen, fmt.Sprintf("puntos %d", j.Player.Points), j.Font.Face, j.Font.Options)
	}
}

func (j *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return j.Dimensiones.Ancho, j.Dimensiones.Alto
}

func main() {
	// cargamos la fuente
	fontFile, err := assetsFS.Open("assets/font.ttf")
	if err != nil {
		log.Fatal(err)
	}

	fontSource, err := text.NewGoTextFaceSource(fontFile)

	if err != nil {
		log.Fatal(err)
	}

	font := &Font{
		Face: &text.GoTextFace{
			Source: fontSource,
			Size:   FontSize,
		},
		Options: &text.DrawOptions{},
	}

	font.Options.ColorScale.ScaleWithColor(color.RGBA{R: 0, G: 255, B: 0, A: 255})
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

	mapa := NewMaze(Columnas, Filas)

	Mazerand(mapa)

	juego := &Game{
		Maze:        mapa,
		Dimensiones: &Dimensiones{},
		Player:      jugador,
		Font:        font,
		MazeAssets: &MazeAssets{
			Floor: openAsset(assetsFS, "assets/floor.png"),
			Wall:  openAsset(assetsFS, "assets/wall.png"),
		},
		State: PlayingState,
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
		NodePosition: NewNode(c-2, f-2), // columnas, filas, se considera que n-1 menos el los muros
		Elapse:       EnemyElapse,       // cada cierto ciclos va recalcular la ruta al enemigo
	}

	juego.Enemy.VectorCurrentPosition = NewVector(
		float64(juego.Enemy.NodePosition.X*squareSize),
		float64(juego.Enemy.NodePosition.Y*squareSize),
	)

	juego.Enemy.Juego = juego

	juego.Enemy.Animation = NewAnimation(&AnimationOption{
		Assets:         assetsFS,
		Indexes:        [2]int{0, 11},
		TemplateString: "assets/dog/f_%d.png",
		Elapse:         TPS * .25,
	})

	// cargamos la animacion de ajolote pesos

	juego.MazeAssets.AjoloteAnimation = NewAnimation(&AnimationOption{
		Assets:         assetsFS,
		Indexes:        [2]int{1, 12},
		TemplateString: "assets/ajolote/f%d.png",
		Elapse:         AjoloteElapse,
	})

	// iniciamos el calculo inicial del enemigo
	juego.Enemy.CalculatePath()

	ebiten.SetWindowSize(juego.Dimensiones.Ancho, juego.Dimensiones.Alto)
	ebiten.SetWindowTitle("Catch me!")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	if err := ebiten.RunGame(juego); err != nil {
		panic(err)
	}
}
