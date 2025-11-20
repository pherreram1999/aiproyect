package main

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type GameScore struct {
	gorm.Model
	Velocity uint      `json:"velocity"`
	Score    uint      `json:"score"`
	Time     time.Time `json:"time"`
}

func OpenDB() (*gorm.DB, error) {
	return gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
}
