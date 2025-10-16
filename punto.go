package main

type Punto struct {
	x, y int
	// para calculo A*
	f      float64 // costo de transitar al nodo
	g      float64 // costo total mas heuristica
	Parent *Punto
}

func NewPoint(x, y int, f, g float64) *Punto {
	return &Punto{
		x: x,
		y: y,
		f: f,
		g: g,
	}
}

func (p *Punto) BuildWay() {

}
