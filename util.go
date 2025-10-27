package main

import "github.com/hajimehoshi/ebiten/v2"

func redimensionarImagen(img *ebiten.Image, anchoDeseado, altoDeseado int) *ebiten.Image {
	// Obtener tamaño original
	bounds := img.Bounds()
	anchoOriginal := bounds.Dx()
	altoOriginal := bounds.Dy()

	// Calcular escalas
	escalaX := float64(anchoDeseado) / float64(anchoOriginal)
	escalaY := float64(altoDeseado) / float64(altoOriginal)

	// Crear nueva imagen del tamaño deseado
	nuevaImagen := ebiten.NewImage(anchoDeseado, altoDeseado)

	// Aplicar transformación
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(escalaX, escalaY)

	// Dibujar imagen escalada
	nuevaImagen.DrawImage(img, op)

	return nuevaImagen
}
