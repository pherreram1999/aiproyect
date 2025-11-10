package main

import (
	"embed"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// ScaleFrame escala una imagen a un tamaño específico
func ScaleFrame(originalFrame *ebiten.Image, targetWidth, targetHeight int) *ebiten.Image {
	bounds := originalFrame.Bounds()
	originalWidth := float64(bounds.Dx())
	originalHeight := float64(bounds.Dy())

	// Calcular escalas
	scaleX := float64(targetWidth) / originalWidth
	scaleY := float64(targetHeight) / originalHeight

	// Crear nueva imagen escalada
	scaledFrame := ebiten.NewImage(targetWidth, targetHeight)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scaleX, scaleY)
	scaledFrame.DrawImage(originalFrame, op)

	return scaledFrame
}

func openAsset(assets embed.FS, assetName string) *ebiten.Image {
	f, err := assets.Open(assetName)
	if err != nil {
		log.Fatal(err)
	}
	frame, _, err := ebitenutil.NewImageFromReader(f)

	if err != nil {
		log.Fatal(err)
	}

	return ScaleFrame(frame, squareSize, squareSize)
}
