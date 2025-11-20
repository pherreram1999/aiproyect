package main

import (
	"fmt"
	"math"
)

type Vector2d struct {
	X, Y float64
}

func NewVector(x, y float64) *Vector2d {
	return &Vector2d{x, y}
}

func (v *Vector2d) Clone() *Vector2d {
	return &Vector2d{v.X, v.Y}
}

func (self *Vector2d) Add(other *Vector2d) *Vector2d {
	return &Vector2d{
		X: self.X + other.X,
		Y: self.Y + other.Y,
	}
}

func (self *Vector2d) Sub(other *Vector2d) *Vector2d {
	return &Vector2d{
		X: self.X - other.X,
		Y: self.Y - other.Y,
	}
}

func (self *Vector2d) SquaredDistance() float64 {
	return self.X*self.X + self.Y*self.Y
}

func (self *Vector2d) Distance() float64 {
	return math.Sqrt(self.SquaredDistance())
}

func (self *Vector2d) Normalize() *Vector2d {
	mod := self.Distance()
	return &Vector2d{
		X: self.X / mod,
		Y: self.Y / mod,
	}
}

func (self *Vector2d) MultiplyByScalar(scalar float64) *Vector2d {
	return &Vector2d{
		X: self.X * scalar,
		Y: self.Y * scalar,
	}
}

func (self *Vector2d) String() string {
	return fmt.Sprintf("(%f.2,%f.2)", self.X, self.Y)
}
