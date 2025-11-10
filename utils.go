package main

import "github.com/hajimehoshi/ebiten/v2"

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
