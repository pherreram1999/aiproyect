package main

import (
	"fmt"
	"math"
)

type Vector struct {
	X, Y float64
}

func NewVector(x, y float64) *Vector {
	return &Vector{x, y}
}

func (self *Vector) Add(other *Vector) *Vector {
	return &Vector{
		X: self.X + other.X,
		Y: self.Y + other.Y,
	}
}

func (self *Vector) Sub(other *Vector) *Vector {
	return &Vector{
		X: self.X - other.X,
		Y: self.Y - other.Y,
	}
}

func (self *Vector) SquaredDistance() float64 {
	return self.X*self.X + self.Y*self.Y
}

func (self *Vector) Distance() float64 {
	return math.Sqrt(self.SquaredDistance())
}

func (self *Vector) Normalize() *Vector {
	mod := self.Distance()
	return &Vector{
		X: self.X / mod,
		Y: self.Y / mod,
	}
}

func (self *Vector) MultiplyByScalar(scalar float64) *Vector {
	return &Vector{
		X: self.X * scalar,
		Y: self.Y * scalar,
	}
}

func (self *Vector) String() string {
	return fmt.Sprintf("(%f.2,%f.2)", self.X, self.Y)
}
