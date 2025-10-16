package main

import (
	"fmt"
	"math"
)

type Vector struct {
	x, y float64
}

func (self *Vector) Add(other *Vector) *Vector {
	return &Vector{
		x: self.x + other.x,
		y: self.y + other.y,
	}
}

func (self *Vector) Sub(other *Vector) *Vector {
	return &Vector{
		x: self.x - other.x,
		y: self.y - other.y,
	}
}

func (self *Vector) SquaredDistance() float64 {
	return self.x*self.x + self.y*self.y
}

func (self *Vector) Distance() float64 {
	return math.Sqrt(self.SquaredDistance())
}

func (self *Vector) Normalize() *Vector {
	mod := self.Distance()
	return &Vector{
		x: self.x / mod,
		y: self.y / mod,
	}
}

func (self *Vector) MultiplyByScalar(scalar float64) *Vector {
	return &Vector{
		x: self.x * scalar,
		y: self.y * scalar,
	}
}

func (self *Vector) String() string {
	return fmt.Sprintf("(%f.2,%f.2)", self.x, self.y)
}
