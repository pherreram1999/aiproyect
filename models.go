package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type GameScore struct {
	gorm.Model
	Velocity int     `json:"velocity"`
	Score    uint    `json:"score"`
	Time     float64 `json:"time"`
}

func OpenDB() (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(DbName), &gorm.Config{})
}
