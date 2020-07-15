package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
)

// Преподаватель
type Teacher struct {
	gorm.Model
	FirstName   string
	MiddleName  string // отчество
	LastName    string
	Disciplines []*Discipline `gorm:"many2many:disciplines_teachers"`
	Cathedras   []*Cathedra   `gorm:"many2many:cathedras_teachers"`
	Groups      []*Group      `gorm:"many2many:groups_teachers"`
}
