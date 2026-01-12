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
		dataY[i] = float64(s.Velocity)
	}

	X := mat.NewDense(numSamples, numFeatures, dataX)
	Yd := mat.NewDense(numSamples, 1, dataY)

	// 4. Normalizar datos
	fmt.Println("Normalizando datos...")
	X_norm, _, _ := NormalizarZScore(X)
	// Opcional: Normalizar Yd también si los valores son muy grandes,
	// pero para regresión/clasificación simple a veces no es estricto.
	// Sin embargo, la red usa sigmoide en la salida (0-1), así que SI debemos escalar Yd si sus valores salen de 0-1.
	// El usuario dijo "extraer velocity como una matriz de Yd", no especificó normalizar,
	// pero la función 'Entrenar' usa sigmoide en la salida final.
	// Si Velocity > 1, la red saturará.
	// Revisando 'Entrenar', la capa de salida tiene sigmoide.
	// Si Velocity es int (ej. 100, 200), necesitamos escalarlo o cambiar la función de activación de salida.
	// Dado el código existente en network.go, 'Entrenar' fuerza sigmoide en todas las capas.
	// Por lo tanto, Yd DEBE estar entre 0 y 1.

	// Vamos a normalizar Yd usando MinMax o dividiendo por un máximo conocido.
	// O usamos ZScore y luego desnormalizamos al predecir.
	// Usaré ZScore para Yd también para ser consistentes con las herramientas que tengo.
	Yd_norm, _, _ := NormalizarZScore(Yd)

	// 5. Configurar red
	n_in := 2
	n_out := 1
	n_layers := []int{4} // Capa oculta con 4 neuronas
	lr := 0.01
	epochs := 50000

	fmt.Println("Entrenando modelo...")
	W, B, ecm := Entrenar(X_norm, Yd_norm, n_in, n_out, n_layers, lr, epochs)

	if len(ecm) > 0 {
		fmt.Printf("Error final: %f\n", ecm[len(ecm)-1])
	}

	// 6. Guardar modelo
	GuardarModelo("modelo_velocidad", W, B)
	fmt.Println("Entrenamiento finalizado.")
}
