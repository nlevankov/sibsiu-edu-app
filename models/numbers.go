package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
)

// Номер зачетки
// Отдельная табилца нужна потому, чтоб понимать, к какой группе относится зачетка,
// а не просто поле типа "массив" у студента с номерами зачеток
type Number struct {
	gorm.Model
	Number    uint `sql:"unique"`
	StudentID uint
	Group     string
}
