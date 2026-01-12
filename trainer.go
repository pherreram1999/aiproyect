package main

import (
	"fmt"
	"log"

	"gonum.org/v1/gonum/mat"
)

func RunTraining() {
	fmt.Println("Iniciando proceso de entrenamiento...")

	// 1. Abrir Base de Datos
	db, err := OpenDB()
	if err != nil {
		log.Fatalf("Error al abrir la base de datos: %v", err)
	}

	// 2. Extraer datos
	var scores []GameScore
	result := db.Find(&scores)
	if result.Error != nil {
		log.Fatalf("Error al leer datos de la BD: %v", result.Error)
	}

	if len(scores) == 0 {
		log.Fatal("No hay datos en la base de datos para entrenar.")
	}

	fmt.Printf("Se encontraron %d registros para entrenamiento.\n", len(scores))

	// 3. Construir matrices X (Inputs) y Yd (Target)
	// Inputs: Score, Time
	// Target: Velocity
	numSamples := len(scores)
	numFeatures := 2 // Score, Time

	dataX := make([]float64, numSamples*numFeatures)
	dataY := make([]float64, numSamples)

	for i, s := range scores {
		// X: [Score, Time]
		dataX[i*numFeatures+0] = float64(s.Score)
		dataX[i*numFeatures+1] = s.Time

		// Yd: [Velocity]
		// Normalizamos MinMax (0-1) para que coincida con Sigmoide
		// Velocity (Elapse) va de Min=30 (Rápido) a Max=80 (Lento)
		// dataY[i] = float64(s.Velocity)
		normY := (float64(s.Velocity) - float64(EnemyElapseMin)) / float64(EnemyElapseMax-EnemyElapseMin)
		if normY < 0 {
			normY = 0
		}
		if normY > 1 {
			normY = 1
		}
		dataY[i] = normY
	}

	X := mat.NewDense(numSamples, numFeatures, dataX)
	Yd := mat.NewDense(numSamples, 1, dataY)

	// 4. Normalizar datos de Entrada (X) con Z-Score
	fmt.Println("Normalizando datos...")
	X_norm, media, desviacion := NormalizarZScore(X)

	// Yd ya está normalizado a 0-1 manualmente

	// 5. Configurar red
	n_in := 2
	n_out := 1
	n_layers := []int{4} // Capa oculta con 4 neuronas
	lr := 0.01
	epochs := 50000

	fmt.Println("Entrenando modelo...")
	W, B, ecm := Entrenar(X_norm, Yd, n_in, n_out, n_layers, lr, epochs)

	if len(ecm) > 0 {
		fmt.Printf("Error final: %f\n", ecm[len(ecm)-1])
	}

	// 6. Desnormalizar para que el modelo acepte inputs crudos
	// Integramos la normalización en la primera capa
	W, B = Desnormalizar(W, B, media, desviacion)

	// 7. Guardar modelo
	GuardarModelo("modelo_velocidad", W, B)
	fmt.Println("Entrenamiento finalizado.")
}
