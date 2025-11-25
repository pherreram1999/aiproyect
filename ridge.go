package main

import (
	"math"
	"math/rand"
)

type Vector []float64

func (v Vector) Sum() float64 {
	sum := 0.0
	for i := 0; i < len(v); i++ {
		sum += v[i]
	}
	return sum
}

func (v Vector) Media() float64 {
	n := len(v)
	if n == 0 {
		return 0.0
	}
	return v.Sum() / float64(n)
}

func (v Vector) StandardDeviation() float64 {
	n := len(v)
	if n == 0 {
		return 0.0
	}
	sumaCuadrica := 0.0
	media := v.Media()
	for i := 0; i < len(v); i++ {
		x := v[i] - media
		sumaCuadrica += x * x
	}
	return math.Sqrt(sumaCuadrica / float64(n))
}

func (v Vector) Normalizar() (Vector, float64, float64) {
	normalizado := make(Vector, len(v))
	media := v.Media()
	desviacion := v.StandardDeviation()

	if desviacion == 0.0 { // todos los valores son iguales, por defecto el arreglo son de ceros
		return normalizado, media, desviacion
	}

	for i := 0; i < len(v); i++ {
		normalizado[i] = (v[i] - media) / desviacion
	}
	return normalizado, media, desviacion
}

func (v Vector) DesNormalizar(media, desviacion float64) Vector {
	desnormalizado := make(Vector, len(v))

	for i := 0; i < len(v); i++ {
		desnormalizado[i] = (v[i] * desviacion) + media
	}

	return desnormalizado
}

func Ridge(Yr Vector, X []Vector) {
	m := float64(len(X)) // numero de filas (datos)
	if m == 0 {
		// regresar el errors
	}
	numCaracteristicas := len(X[0]) // numero de caracteristicas columnas
	betas := make(Vector, numCaracteristicas)

	b0 := -rand.Float64() // se recomiendas que la primera beta sea negativa
	// incializamos las otras betas
	for i := range betas {
		betas[i] = rand.Float64()
	}

	epilon := 0.000001
	MaximoEpocas := 150_000
	epocas := 0
	lr := 0.0072   // learning rate
	T := 0.0000001 // lambda de penilizacion

	// J el nuestro error de modelo, tratamos de hacerlo lo mas chica posible
	var Jactual float64 = 0
	var Jant float64 = 10

	// vamos a seguir iterando hasta que la diferencia del costo con su anterior se menor que epsilon
	// al usar abs solo nos interesa el cambi, en alguno punto, mientras siga siendo mayor, se sigue aprendiendo
	// si llega algo menor que epsion, ya se detiene, ha convergido
	for epocas < MaximoEpocas && math.Abs(Jactual-Jant) > epilon {
		// hacer el calculo de modelo propuesto

		// almacenamos las diferencias de Yr y Ym para el calculo de J
		var sumaCuadradaYDelta float64
		// para el calculo de error de b0
		var sumaYDelta float64
		// almacenamos los gradientes de betas
		gradientes := make(Vector, numCaracteristicas)
		for ix, fx := range X {
			// calcular el modelo propuesto
			// Ym = b0 + bnXn
			Ym := b0
			for jx, valorX := range fx {
				Ym += valorX * betas[jx]
			}

			// revisamos que tan lejos esta respecto a nuestra Y real
			deltaY := Ym - Yr[ix]
			sumaYDelta += deltaY
			sumaCuadradaYDelta += deltaY * deltaY
			// calculamos lo gradientes individuales de cada beta

			// es producto punto de deltaY (el error) por cada valor de
			for jx, valorX := range fx { // por cada valor de fx (la fila actual, que es nuestra caracteristica actual
				// se multiplca por deltaY para por cada valor de nuestra caracteristica actual
				// esto para obtener el vector gradiente
				gradientes[jx] += deltaY * valorX
			}
		}

		Jant = Jactual

		// aplicamos el cuadrado al
		var sumaBetasCuadras float64

		for _, b := range betas {
			sumaBetasCuadras += b * b
		}

		Jactual = (1.0/(2.0*m))*sumaCuadradaYDelta + T*sumaBetasCuadras

		// actualizamos las betas
		b0 = b0 - (lr/m)*sumaYDelta

		for jg := 0; jg < numCaracteristicas; jg++ {
			betas[jg] = betas[jg] - lr*((gradientes[jg]/m)+(T*betas[jg]))
		}
		epocas++
	}

}
