package main

import (
	"encoding/gob"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"gonum.org/v1/gonum/mat"
)

// Inicializamos la semilla aleatoria globalmente
func init() {
	rand.Seed(time.Now().UnixNano())
}

// ==========================================
// FUNCIONES AUXILIARES MATEMÁTICAS
// ==========================================

// sigmoide aplica la función sigmoide a un valor escalar.
// Equivale a: 1/(1+np.exp(-z))
func sigmoide(z float64) float64 {
	return 1.0 / (1.0 + math.Exp(-z))
}

// d_sigmoide_z calcula la derivada de la sigmoide basada en Z.
// Equivale a: s*(1-s)
func d_sigmoide_z(z float64) float64 {
	s := sigmoide(z)
	return s * (1.0 - s)
}

// applySigmoide aplica la sigmoide a cada elemento de una matriz.
// Numpy lo hace automáticamente (vectorización), en Gonum usamos Apply.
func applySigmoide(m mat.Matrix) *mat.Dense {
	r, c := m.Dims()
	res := mat.NewDense(r, c, nil)
	res.Apply(func(i, j int, v float64) float64 {
		return sigmoide(v)
	}, m)
	return res
}

// applyDSigmoide aplica la derivada de la sigmoide a cada elemento.
func applyDSigmoide(m mat.Matrix) *mat.Dense {
	r, c := m.Dims()
	res := mat.NewDense(r, c, nil)
	res.Apply(func(i, j int, v float64) float64 {
		return d_sigmoide_z(v)
	}, m)
	return res
}

// ==========================================
// LÓGICA PRINCIPAL
// ==========================================

/*
Entrena el modelo segun los patrones dados
:param X: matriz de caractericas de forma (muestras,caracteristicas)
:param Yd: matriz de etiquetas
:param n_in: numero de neuronas entradas
:param n_out: numero de neuronas de salida
:param n_layers: lista de numero de neuronas ocultas por capa
:param lr: learning rate
:param epoch_max: numero de epocas maximas
:return: W (pesos), B (bias), ECM_historico
*/
func Entrenar(X, Yd *mat.Dense, n_in, n_out int, n_layers []int, lr float64, epoch_max int) ([]*mat.Dense, []*mat.Dense, []float64) {

	// contiene cuantas neuronas tiene cada capa
	// [GO vs PY]: En Go concatenamos slices manualmente.
	dimensiones := append([]int{n_in}, n_layers...)
	dimensiones = append(dimensiones, n_out)

	numero_dimensiones := len(dimensiones)
	// numero_capas := len(dimensiones) // No se usa en el original explícitamente

	// slice de matrices
	W := make([]*mat.Dense, 0) // almacena los pesos
	B := make([]*mat.Dense, 0) // almacena las bias

	// incializamos los pesos
	for i := 0; i < numero_dimensiones-1; i++ {
		filas := dimensiones[i]
		cols := dimensiones[i+1] // recordar que es para la matriz siguiente, las operaciones entre matirces

		// [GO vs PY]: np.random.randn(f, c) * 0.1
		// En Gonum creamos la matriz y llenamos los datos manualmente con rand.NormFloat64
		// se incializa los pesos con valores chicos random a un inicio
		wData := make([]float64, filas*cols)
		for k := range wData { // aqui tenemos que hacer fila por fila
			wData[k] = rand.NormFloat64() * 0.1
		}
		W = append(W, mat.NewDense(filas, cols, wData))

		// El bias el valor que nos ayuda incializar los valores de la capa destino
		// [GO vs PY]: np.random.randn(c) * 0.1
		// NOTA: Para facilitar operaciones matriciales en Go, tratamos B como una matriz fila (1, cols)
		bData := make([]float64, cols)
		for k := range bData {
			bData[k] = rand.NormFloat64() * 0.1
		}
		B = append(B, mat.NewDense(1, cols, bData))
	}

	ECM := 10.0 // nuestro error
	ECM_historico := []float64{}
	epoch := 0

	// Obtener número de muestras (filas de X)
	nMuestras, _ := X.Dims()

	for ECM > 0.0 && epoch <= epoch_max {
		suma_ecm := 0.0

		for p := 0; p < nMuestras; p++ { // recorremos los patrones

			// ===== Propagacion hacia adelante

			// [GO vs PY]: A = [X[p]]. En Gonum extraemos la fila 'p' como una nueva matriz (View/Slice).
			// Slice(i, k, j, l) toma filas [i,k) y columnas [j,l).
			rowX := X.Slice(p, p+1, 0, n_in).(*mat.Dense)

			A := []*mat.Dense{rowX} // primer entrada como valores de activacion
			Z := []*mat.Dense{}     // guardamos los valores antes de activar

			for i := 0; i < len(W); i++ {
				// recordar que es cada valor de activacion por cada peso
				// [GO vs PY]: z_actual = np.dot(A[i], W[i]) + B[i]
				// En Gonum: Mul realiza la multiplicación matricial.
				zActual := mat.NewDense(1, B[i].RawMatrix().Cols, nil)
				zActual.Mul(A[i], W[i]) // producto punto

				// [GO vs PY]: Suma de Bias. A diferencia de numpy que hace broadcasting automático,
				// aquí usamos Add. Como definimos B como (1, cols), las dimensiones coinciden.
				zActual.Add(zActual, B[i]) // es deplazamiento de la funcion

				Z = append(Z, zActual)
				A = append(A, applySigmoide(zActual))
			}

			// la ultima activacion es la salida de esperada de Y obtenida
			Y_obt := A[len(A)-1]

			// guardamos los errores para poder propagar hacia atras
			deltas := make([]*mat.Dense, len(W))

			// recordar que devuelve un arreglo de los errores (distancias)
			// [GO vs PY]: error = Yd[p] - Y_obt
			yDeseada := Yd.Slice(p, p+1, 0, n_out).(*mat.Dense)
			errorMat := mat.NewDense(1, n_out, nil)
			errorMat.Sub(yDeseada, Y_obt)

			// [GO vs PY]: suma_ecm += np.sum(error**2)
			// Iteramos manualmente para sumar los cuadrados
			er, ec := errorMat.Dims()
			for r := 0; r < er; r++ {
				for c := 0; c < ec; c++ {
					val := errorMat.At(r, c)
					suma_ecm += val * val
				}
			}

			// la (Deseado - Obtenido) * derivada que conecta los errores entre capas
			// [GO vs PY]: deltas[-1] = error * d_sigmoide_z(Z[-1])
			// MulElem es multiplicación elemento a elemento (Hadamard product), no matricial.
			dSigZ := applyDSigmoide(Z[len(Z)-1])
			deltaLast := mat.NewDense(1, n_out, nil)
			deltaLast.MulElem(errorMat, dSigZ)
			deltas[len(deltas)-1] = deltaLast

			// propagamos el error
			// [GO vs PY]: reversed(range(len(deltas) - 1))
			for i := len(deltas) - 2; i >= 0; i-- {
				// [GO vs PY]: np.dot(deltas[i+1], W[i+1].T)
				// Necesitamos la transpuesta de W[i+1].
				wNextT := W[i+1].T()
				deltaProp := mat.NewDense(1, dimensiones[i+1], nil)
				deltaProp.Mul(deltas[i+1], wNextT)

				// [GO vs PY]: ... * d_sigmoide_z(Z[i])
				// Multiplicación elemento a elemento
				dSigZi := applyDSigmoide(Z[i])
				deltaCurrent := mat.NewDense(1, dimensiones[i+1], nil)
				deltaCurrent.MulElem(deltaProp, dSigZi)

				deltas[i] = deltaCurrent
			}

			// actualizamos pesos
			for i := 0; i < len(W); i++ {
				// [GO vs PY]: W[i] += lr * np.outer(A[i], deltas[i])
				// np.outer(A[i], deltas[i]) donde A[i] es (1, N) y deltas[i] es (1, M) en Numpy
				// produce una matriz (N, M).
				// En Gonum: A[i] es (1, N). Transponemos A[i] a (N, 1).
				// (N, 1) x (1, M) = (N, M).
				changeW := mat.NewDense(dimensiones[i], dimensiones[i+1], nil)
				changeW.Mul(A[i].T(), deltas[i])
				changeW.Scale(lr, changeW) // Multiplicar por learning rate

				W[i].Add(W[i], changeW)

				// [GO vs PY]: B[i] += lr * deltas[i]
				changeB := mat.NewDense(1, dimensiones[i+1], nil)
				changeB.Scale(lr, deltas[i])
				B[i].Add(B[i], changeB)
			}
		}

		// [GO vs PY]: ECM = 0.5 * (suma_ecm / len(X))
		ECM = 0.5 * (suma_ecm / float64(nMuestras))
		ECM_historico = append(ECM_historico, ECM)
		epoch++
	}

	return W, B, ECM_historico
}

/*
Normaliza una matriz de datos usando Z-Score (estandarización).
:param X: Matriz de datos (muestras, caracteristicas)
:return: Matriz normalizada, vector media, vector desviacion
*/
func NormalizarZScore(X *mat.Dense) (*mat.Dense, []float64, []float64) {
	r, c := X.Dims()
	media := make([]float64, c)
	desviacion := make([]float64, c)
	X_norm := mat.NewDense(r, c, nil)

	// [GO vs PY]: np.mean(X, axis=0) y np.std(X, axis=0)
	// En Gonum iteramos columnas manualmente para calcular estadísticas
	for j := 0; j < c; j++ {
		// Extraer columna j
		col := mat.Col(nil, j, X)

		// Calcular media
		sum := 0.0
		for _, v := range col {
			sum += v
		}
		media[j] = sum / float64(r)

		// Calcular desviación
		sumSq := 0.0
		for _, v := range col {
			diff := v - media[j]
			sumSq += diff * diff
		}
		desviacion[j] = math.Sqrt(sumSq / float64(r)) // numpy std usa N por defecto, no N-1

		// [GO vs PY]: desviacion[desviacion == 0] = 1.0
		if desviacion[j] == 0 {
			desviacion[j] = 1.0
		}
	}

	// [GO vs PY]: (X - media) / desviacion (Broadcasting)
	// En Gonum usamos Apply para iterar sobre cada celda y aplicar la fórmula
	X_norm.Apply(func(i, j int, v float64) float64 {
		return (v - media[j]) / desviacion[j]
	}, X)

	return X_norm, media, desviacion
}

/*
Ajusta los pesos y bias de la primera capa para aceptar datos sin normalizar.
*/
func Desnormalizar(W []*mat.Dense, B []*mat.Dense, media, desviacion []float64) ([]*mat.Dense, []*mat.Dense) {
	// [GO vs PY]: w.copy().
	// En Gonum usamos CloneFrom para hacer una copia profunda (deep copy)
	W_real := make([]*mat.Dense, len(W))
	B_real := make([]*mat.Dense, len(B))
	for i := range W {
		wr := mat.NewDense(0, 0, nil)
		wr.CloneFrom(W[i])
		W_real[i] = wr

		br := mat.NewDense(0, 0, nil)
		br.CloneFrom(B[i])
		B_real[i] = br
	}

	// Ajustamos solo la PRIMERA capa (índice 0)
	// W_real[0] tiene forma (caracteristicas, neuronas_ocultas)
	rows, cols := W_real[0].Dims()

	// 1. Ajustar Pesos: W / sigma (columna por columna)
	// [GO vs PY]: W_real[0][:, i] = W_real[0][:, i] / desviacion
	// Nota: En la versión python divide la columna i de W por el vector desviacion (broadcasting sobre filas).
	// Matemáticamente: W_new[j, i] = W_old[j, i] / sigma[j] (donde j es la característica de entrada)
	for j := 0; j < rows; j++ { // j itera sobre características (filas de W)
		for i := 0; i < cols; i++ { // i itera sobre neuronas ocultas (columnas de W)
			// Dividimos el peso que conecta la entrada j con la neurona i por la std de la entrada j
			val := W_real[0].At(j, i)
			W_real[0].Set(j, i, val/desviacion[j])
		}
	}

	// 2. Ajustar Bias: B - sum(mu * W / sigma)
	// [GO vs PY]: ajuste_bias = np.dot(media, W_real[0])
	// media es (1, n_features), W_real[0] es (n_features, n_hidden).
	mediaVec := mat.NewDense(1, len(media), media)
	ajusteBias := mat.NewDense(1, cols, nil)
	ajusteBias.Mul(mediaVec, W_real[0])

	// [GO vs PY]: B_real[0] = B_real[0] - ajuste_bias
	B_real[0].Sub(B_real[0], ajusteBias)

	return W_real, B_real
}

func Predecir(x *mat.Dense, W []*mat.Dense, B []*mat.Dense) *mat.Dense {
	activacion := x
	for i := 0; i < len(W); i++ {
		// [GO vs PY]: z = np.dot(activacion, W[i]) + B[i]
		_, cW := W[i].Dims()
		z := mat.NewDense(1, cW, nil)
		z.Mul(activacion, W[i])
		z.Add(z, B[i])

		// La nueva activación es el resultado de la sigmoide
		activacion = applySigmoide(z)
	}
	return activacion
}

// Escalonar convierte probabilidades en 0 o 1
func Escalonar(Z *mat.Dense) []float64 {
	_, cols := Z.Dims()
	res := make([]float64, cols)
	for j := 0; j < cols; j++ {
		val := Z.At(0, j)
		if val > 0.5 {
			res[j] = 1.0
		} else {
			res[j] = 0.0
		}
	}
	return res
}

// Estructura auxiliar para guardar en archivo (Gob)
type ModelDump struct {
	WData [][]float64
	WRows []int
	WCols []int
	BData [][]float64
	BRows []int
	BCols []int
}

func GuardarModelo(nombreModelo string, W []*mat.Dense, B []*mat.Dense) {
	// En Go no existe np.savez directo, usamos encoding/gob para serializar
	dump := ModelDump{}

	for _, w := range W {
		r, c := w.Dims()
		dump.WRows = append(dump.WRows, r)
		dump.WCols = append(dump.WCols, c)
		dump.WData = append(dump.WData, mat.Col(nil, 0, w.T().(*mat.Dense))) // Flatten data
	}
	for _, b := range B {
		r, c := b.Dims()
		dump.BRows = append(dump.BRows, r)
		dump.BCols = append(dump.BCols, c)
		dump.BData = append(dump.BData, mat.Col(nil, 0, b.T().(*mat.Dense)))
	}

	file, err := os.Create(nombreModelo + ".gob")
	if err != nil {
		fmt.Println("Error creando archivo:", err)
		return
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	encoder.Encode(dump)
	fmt.Println("El calculo de W y B se guardo en disco: " + nombreModelo + ".gob !!!")
}

/*
CargarModelo lee un archivo .gob y reconstruye las matrices de pesos y bias.
:param nombreModelo: nombre del archivo (sin extensión .gob)
:return: W ([]*mat.Dense), B ([]*mat.Dense), error
*/
func CargarModelo(nombreModelo string) ([]*mat.Dense, []*mat.Dense, error) {
	// 1. Abrir el archivo
	file, err := os.Open(nombreModelo + ".gob")
	if err != nil {
		return nil, nil, fmt.Errorf("no se pudo abrir el archivo: %v", err)
	}
	defer file.Close()

	// 2. Decodificar la estructura ModelDump
	var dump ModelDump
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&dump)
	if err != nil {
		return nil, nil, fmt.Errorf("error al decodificar el modelo: %v", err)
	}

	// 3. Reconstruir los Pesos (W)
	W := make([]*mat.Dense, len(dump.WRows))
	for i := 0; i < len(dump.WRows); i++ {
		// NewDense toma (filas, columnas, data)
		// Gonum copiará el slice dump.WData[i] para crear la matriz
		W[i] = mat.NewDense(dump.WRows[i], dump.WCols[i], dump.WData[i])
	}

	// 4. Reconstruir los Bias (B)
	B := make([]*mat.Dense, len(dump.BRows))
	for i := 0; i < len(dump.BRows); i++ {
		B[i] = mat.NewDense(dump.BRows[i], dump.BCols[i], dump.BData[i])
	}

	fmt.Printf("Modelo '%s.gob' cargado correctamente.\n", nombreModelo)
	return W, B, nil
}
