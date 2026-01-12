package main

import (
	"embed"
	"fmt"
	"image/color"
	_ "image/png"
	"io"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"gorm.io/gorm"
)

type State int

const (
	MenuState State = iota
	PlayingState
	GameOverState
	ScoresState
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
	MenuIndex     int
	Scores        []GameScore
	ScoresLoaded  bool
	LastScore     uint
	LastTimeSec   float64
	LastVelocity  int
	Maze          Maze         // guarda la matriz del mapa del juego
	Dimensiones   *Dimensiones // guarda las dimensiones del mapa del juego
	Player        *Player      // guarda el objeto del datos del jugador
	IsMoving      bool         // indica si se esta moviendo en el mapa
	MazeAssets    *MazeAssets  // contiene texturas para el renderizado del mapa
	Enemys        Enemys
	State         State     // indica el estado actual del juego, si esta jugado o ha terminad
	Font          *Font     // fuente para renderizar en el juego
	StartTime     time.Time // indica cuando empezo la partida del jugador
	DB            *gorm.DB
	CoinSoundData []byte
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
		// sonamos que tomo una ajolote point
		j.playCoinSound()
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
	diff := time.Since(j.StartTime).Seconds()
	e := j.Enemys[0]

	// Guardar “último resultado” para mostrarlo en pantalla
	j.LastScore = j.Player.Points
	j.LastTimeSec = diff
	j.LastVelocity = e.Elapse

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

// Funciones de Menu
func (j *Game) resetRun() {
	// Reinicio mínimo para “Jugar” SIN rearmar todo el juego ni tocar tu lógica.
	// Esto evita romper el estado y mantiene tu estructura.
	j.Player.Points = 0
	j.StartTime = time.Now()
	j.State = PlayingState
}

func (j *Game) loadTopScores() {
	if j.ScoresLoaded {
		return
	}
	var scores []GameScore
	// Top 10 por Score DESC (si quieres desempate por tiempo, lo puedes ajustar)
	err := j.DB.Order("score DESC").Limit(10).Find(&scores).Error
	if err != nil {
		// No reventamos el juego; solo dejamos vacío
		j.Scores = nil
		j.ScoresLoaded = true
		return
	}
	j.Scores = scores
	j.ScoresLoaded = true
}

func (j *Game) UpdateMenu() error {
	// Navegación (una vez por pulsación)
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		j.MenuIndex--
		if j.MenuIndex < 0 {
			j.MenuIndex = 2
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		j.MenuIndex++
		if j.MenuIndex > 2 {
			j.MenuIndex = 0
		}
	}

	// Selección
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		switch j.MenuIndex {
		case 0: // Jugar
			j.StartNewRun()
		case 1: // Puntuaciones
			j.ScoresLoaded = false
			j.loadTopScores()
			j.State = ScoresState
		case 2: // Salir
			return ebiten.Termination
		}
	}

	return nil
}

func (j *Game) UpdateScores() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		j.State = MenuState
	}
	return nil
}

func (j *Game) DrawMenu(screen *ebiten.Image) {
	screen.Fill(color.Black)

	scale := 1.0

	// ---------- TÍTULO ----------
	title := "Catch me!"
	w, _ := text.Measure(title, j.Font.Face, scale)

	titleX := (float64(j.Dimensiones.Ancho) - w) / 2
	titleY := 80.0

	j.Font.Options.GeoM.Translate(titleX, titleY)
	text.Draw(screen, title, j.Font.Face, j.Font.Options)
	j.Font.Options.GeoM.Reset()

	// ---------- MENÚ ----------
	opts := []string{"Jugar", "Puntuaciones", "Salir"}

	startY := float64(j.Dimensiones.Alto)/2 - float64(len(opts)*40)/2

	for i, o := range opts {
		prefix := "  "
		if i == j.MenuIndex {
			prefix = "> "
		}

		line := prefix + o
		lw, _ := text.Measure(line, j.Font.Face, scale)

		x := (float64(j.Dimensiones.Ancho) - lw) / 2
		y := startY + float64(i*40)

		j.Font.Options.GeoM.Translate(x, y)
		text.Draw(screen, line, j.Font.Face, j.Font.Options)
		j.Font.Options.GeoM.Reset()
	}
}

func (j *Game) DrawScores(screen *ebiten.Image) {
	screen.Fill(color.Black)

	text.Draw(screen, "Puntuaciones (Top 10)", j.Font.Face, j.Font.Options)
	j.Font.Options.GeoM.Translate(0, 50)

	if len(j.Scores) == 0 {
		text.Draw(screen, "Sin registros.", j.Font.Face, j.Font.Options)
		j.Font.Options.GeoM.Reset()
		return
	}

	// Lista
	y := 0
	for i, s := range j.Scores {
		line := fmt.Sprintf("%2d) Score: %d  Time: %.2fs  Vel: %d", i+1, s.Score, s.Time, s.Velocity)

		// Dibujar línea en Y
		j.Font.Options.GeoM.Translate(0, float64(y))
		text.Draw(screen, line, j.Font.Face, j.Font.Options)

		// regresar X y avanzar Y
		j.Font.Options.GeoM.Translate(0, -float64(y))
		y += 30
	}

	// Footer
	j.Font.Options.GeoM.Translate(0, float64(50+y))
	text.Draw(screen, "ESC o ENTER para volver", j.Font.Face, j.Font.Options)

	j.Font.Options.GeoM.Reset()
}

func (j *Game) StartNewRun() {
	// 1) Reiniciar jugador
	puntoInicial := NewNode(1, 1)
	middleX := float64(puntoInicial.X * squareSize)
	middleY := float64(puntoInicial.Y * squareSize)

	startPosition := NewVector(middleX, middleY)
	j.Player.Points = 0
	j.Player.IsMoving = false
	j.Player.CurrentPosition = startPosition.Clone()
	j.Player.TargetPosition = startPosition.Clone()
	j.Player.NodePosition = NewNode(1, 1)

	// 2) Nuevo mapa
	mapa := NewMaze(j.Dimensiones.Columnas, j.Dimensiones.Filas)
	Mazerand(mapa)
	j.Maze = mapa

	// 3) Reiniciar enemigos (limpio y seguro)
	j.Enemys = nil

	f := j.Dimensiones.Filas
	c := j.Dimensiones.Columnas

	delta := EnemyElapseMax - EnemyElapseMin
	pasos := MaxAjolotePoints / delta
	deltaStep := delta / pasos

	j.NewEnemy(NewNode(c-2, f-2), deltaStep)
	j.NewEnemy(NewNode(c-2, 1), deltaStep)
	j.NewEnemy(NewNode(1, f-2), deltaStep)

	for _, e := range j.Enemys {
		e.CalculatePath()
	}

	// 4) Reiniciar cronómetro
	j.StartTime = time.Now()

	// 5) A jugar
	j.State = PlayingState
}

//Fin de Menu

func (j *Game) Update() error {
	switch j.State {
	case MenuState:
		return j.UpdateMenu()

	case ScoresState:
		return j.UpdateScores()

	case GameOverState:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			j.StartNewRun()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			j.State = MenuState
		}

	case PlayingState:
		j.MovePlayer()
		j.MoveEnemy()
		for _, e := range j.Enemys {
			if e.NodePosition.Equal(j.Player.NodePosition) || j.Player.Points == MaxAjolotePoints {
				j.GameOver()
			}
		}
	}
	return nil
}

func (j *Game) Draw(screen *ebiten.Image) {
	switch j.State {
	case MenuState:
		j.DrawMenu(screen)
		return

	case ScoresState:
		j.DrawScores(screen)
		return

	case PlayingState:
		j.DrawMaze(screen)

		j.Player.DrawPlayer(screen)
		for _, e := range j.Enemys {
			e.Draw(screen)
		}

		j.MazeAssets.AjoloteAnimation.Tick()

		text.Draw(screen, fmt.Sprintf("Puntaje: %d", j.Player.Points), j.Font.Face, j.Font.Options)
		return

	case GameOverState:
		screen.Fill(color.Black)

		text.Draw(screen, "Perdiste", j.Font.Face, j.Font.Options)
		j.Font.Options.GeoM.Translate(0, 40)

		text.Draw(screen, fmt.Sprintf("Puntaje: %d", j.LastScore), j.Font.Face, j.Font.Options)
		j.Font.Options.GeoM.Translate(0, 35)

		text.Draw(screen, fmt.Sprintf("Tiempo: %.2fs", j.LastTimeSec), j.Font.Face, j.Font.Options)
		j.Font.Options.GeoM.Translate(0, 35)

		text.Draw(screen, fmt.Sprintf("Velocidad: %d", j.LastVelocity), j.Font.Face, j.Font.Options)
		j.Font.Options.GeoM.Translate(0, 50)

		text.Draw(screen, "ENTER: jugar de nuevo   ESC: menu", j.Font.Face, j.Font.Options)

		j.Font.Options.GeoM.Reset()
	}
}

func (j *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return j.Dimensiones.Ancho, j.Dimensiones.Alto
}

func PlayMusic() {
	// 1. Obtener el contexto global (NO usar audio.NewContext)
	// Si el contexto no existe, Ebitengine lo inicializa implícitamente aquí.
	ctx := audio.CurrentContext()
	if ctx == nil {
		// Si es nil, lo inicializamos explícitamente (seguridad para versiones muy nuevas)
		ctx = audio.NewContext(sampleRate)
	}

	// 2. Decodificar el archivo (Usando Vorbis/OGG para mejor loop)
	// DecodeWithoutResampling es más eficiente si tu audio ya está en 44100Hz
	soundFile, err := assetsFS.Open("assets/sounds/soundtrack.ogg")
	if err != nil {
		log.Fatal(err)
	}
	d, err := vorbis.DecodeWithSampleRate(sampleRate, soundFile)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Crear el bucle (La forma explícita y moderna)
	// NewInfiniteLoopWithIntro(fuente, bytesIntro, bytesLoop)
	// Para un bucle completo de toda la canción:
	// Intro = 0, LoopLength = Longitud total del archivo decodificado
	loop := audio.NewInfiniteLoopWithIntro(d, 0, d.Length())

	// 4. Crear el reproductor
	player, err := ctx.NewPlayer(loop)
	if err != nil {
		log.Fatal(err)
	}

	player.SetVolume(0.5)
	player.Play()
}

// Función auxiliar para reproducir el sonido
func (g *Game) playCoinSound() {
	// 3. Crear un reproductor "desechable" desde la memoria
	// NewPlayerFromBytes es muy ligero y eficiente para SFX
	ctx := audio.CurrentContext()
	if ctx == nil {
		ctx = audio.NewContext(sampleRate)
	}

	player := ctx.NewPlayerFromBytes(g.CoinSoundData)

	// Opcional: Variar un poco el volumen o añadir lógica extra
	player.SetVolume(0.8)

	// Reproducir
	player.Play()
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
			Size:   FontSize + 8,
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
		State:     MenuState,
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

	PlayMusic()

	// cargamos el sonido de la moneda

	coinFile, err := assetsFS.Open("assets/sounds/coin.ogg")
	if err != nil {
		log.Fatal(err)
	}
	d, err := vorbis.DecodeWithSampleRate(sampleRate, coinFile)
	if err != nil {
		log.Fatal(err)
	}

	juego.CoinSoundData, err = io.ReadAll(d)

	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowSize(juego.Dimensiones.Ancho, juego.Dimensiones.Alto)
	ebiten.SetWindowTitle("Catch me!")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	if err := ebiten.RunGame(juego); err != nil {
		panic(err)
	}
}
