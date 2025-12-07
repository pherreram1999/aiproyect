package main

import (
	"embed"
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"gorm.io/gorm"
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
	Maze        Maze         // guarda la matriz del mapa del juego
	Dimensiones *Dimensiones // guarda las dimensiones del mapa del juego
	Player      *Player      // guarda el objeto del datos del jugador
	IsMoving    bool         // indica si se esta moviendo en el mapa
	MazeAssets  *MazeAssets  // contiene texturas para el renderizado del mapa
	Enemys      Enemys
	State       State     // indica el estado actual del juego, si esta jugado o ha terminad
	Font        *Font     // fuente para renderizar en el juego
	StartTime   time.Time // indica cuando empezo la partida del jugador
	DB          *gorm.DB
}

func (j *Game) NewEnemy(position *Node, elapse int) {
	e := &Enemy{
		NodePosition:    position,       // columnas, filas, se considera que n-1 menos el los muros
		Elapse:          EnemyElapseMax, // cada cierto ciclos va recalcular la ruta al enemigo
		ElapseDecrement: elapse,         // cada punto cuesta un una parte del recorrido
		PathIndex:       1,
	}

	e.Animation = NewAnimation(&AnimationOption{
		Assets:         assetsFS,
		Indexes:        [2]int{0, 11},
		TemplateString: "assets/dog/f_%d.png",
		Elapse:         TPS * .25,
	})

	e.VectorCurrentPosition = NewVector(
		float64(e.NodePosition.X*squareSize),
		float64(e.NodePosition.Y*squareSize),
	)

	e.Juego = j
	j.Enemys = append(j.Enemys, e)
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
		// ajustamos el intervalor de tiempo para aumentar dificultad
		for _, e := range j.Enemys {
			e.Elapse -= e.ElapseDecrement
		}
		j.Maze.Set(j.Player.NodePosition.X, j.Player.NodePosition.Y, Transitable) // indicamos que ya solo es camino
	}
}

func (j *Game) MoveEnemy() {
	//e := j.Enemy
	//e.Tick() // avanzar animaciones del enemigo
	for _, e := range j.Enemys {
		e.Tick()
	}
}

func (j *Game) GameOver() {
	// tenemos que registrar el puntaje del jugados
	diff := time.Now().Sub(j.StartTime).Seconds()
	e := j.Enemys[0] // tomamos el primer enemigo, todos comparten el mismo dato
	err := j.DB.Create(&GameScore{
		Velocity: e.Elapse,
		Score:    j.Player.Points,
		Time:     diff,
	}).Error

	if err != nil {
		log.Fatal(err)
	}

	j.State = GameOverState

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

		for _, e := range j.Enemys { // validamos si alguno de los enemigos toca al jugador
			if e.NodePosition.Equal(j.Player.NodePosition) || j.Player.Points == MaxAjolotePoints {
				// indicamos que tenemos que acabar el juego
				j.GameOver()
			}
		}

	}

	return nil
}

func (j *Game) Draw(screen *ebiten.Image) {

	// dibujamos le puntaje
	text.Draw(screen, fmt.Sprintf("puntos %d", j.Player.Points), j.Font.Face, j.Font.Options)
	if j.State == PlayingState {
		j.DrawMaze(screen)
		// dibujamos el jugaodor
		// lo colocamos en medio de la celda
		j.Player.DrawPlayer(screen)

		//j.Enemy.Draw(screen)

		for _, e := range j.Enemys {
			e.Draw(screen)
		}

		// animacion para los ajolote poins
		j.MazeAssets.AjoloteAnimation.Tick()

	} else if j.State == GameOverState {
		// dibujamos el games over

	}
}

func (j *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return j.Dimensiones.Ancho, j.Dimensiones.Alto
}

func main() {
	// incializamos la base datos

	db, err := OpenDB()

	if err != nil {
		log.Fatal(err)
	}

	if err = db.AutoMigrate(&GameScore{}); err != nil {
		log.Fatal(err)
	}

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
		State:     PlayingState,
		DB:        db,
		StartTime: time.Now(),
	}

	// para que el jugador tenga acceso al los datos del juego
	juego.Player.Game = juego

	f, c := juego.Maze.GetShape()

	juego.Dimensiones.Alto = f * squareSize
	juego.Dimensiones.Ancho = c * squareSize
	juego.Dimensiones.Filas = f
	juego.Dimensiones.Columnas = c

	// cargamos al enemigo

	delta := EnemyElapseMax - EnemyElapseMin // recorrido en minimo y maximo (distancia)
	pasos := MaxAjolotePoints / delta        // cuantos pasos hay el recorrido, segun cuantos puntos maximos halla

	deltaStep := delta / pasos

	juego.NewEnemy(NewNode(c-2, f-2), deltaStep)
	juego.NewEnemy(NewNode(c-2, 1), deltaStep)
	juego.NewEnemy(NewNode(1, f-2), deltaStep)

	// cargamos la animacion de ajolote pesos
	juego.MazeAssets.AjoloteAnimation = NewAnimation(&AnimationOption{
		Assets:         assetsFS,
		Indexes:        [2]int{1, 12},
		TemplateString: "assets/ajolote/f%d.png",
		Elapse:         AjoloteElapse,
	})
	// iniciamos el calculo inicial del enemigo
	for _, e := range juego.Enemys {
		e.CalculatePath()
	}
	//juego.Enemy.CalculatePath()

	ebiten.SetWindowSize(juego.Dimensiones.Ancho, juego.Dimensiones.Alto)
	ebiten.SetWindowTitle("Catch me!")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	if err := ebiten.RunGame(juego); err != nil {
		panic(err)
	}
}
