package main

import (
	"embed"
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Animation struct {
	Frames      []*ebiten.Image
	NoFrames    int
	Elapse      int
	Index       int
	TickCounter int
}

type AnimationOption struct {
	Assets         embed.FS
	TemplateString string
	Indexes        [2]int
	Elapse         int
}

func NewAnimation(options *AnimationOption) *Animation {
	// cargamos los frames
	a := &Animation{
		Elapse: options.Elapse,
	}

	for i := options.Indexes[0]; i < options.Indexes[1]; i++ {
		f, err := options.Assets.Open(
			fmt.Sprintf(options.TemplateString, i),
		)
		if err != nil {
			log.Fatal(err)
		}

		frame, _, err := ebitenutil.NewImageFromReader(f)

		if err != nil {
			log.Fatal(err)
		}

		a.Frames = append(a.Frames, frame)
	}

	a.NoFrames = len(a.Frames)

	return a
}

// Tick avance tick para las animaciones
func (a *Animation) Tick() {
	a.TickCounter++
	if a.TickCounter > a.Elapse {
		a.TickCounter = 0 // reniciamos el contador
		a.Index++
		if a.Index >= a.NoFrames {
			a.Index = 0 // reniciamos frames
		}
	}
}

func (a *Animation) GetFrame() *ebiten.Image {
	return a.Frames[a.Index]
}
